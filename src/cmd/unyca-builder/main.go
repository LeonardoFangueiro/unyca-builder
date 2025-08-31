package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"unyca-builder/src/internal/ansible"
	"unyca-builder/src/internal/schema"
	"unyca-builder/src/internal/snapshots"
	"unyca-builder/src/internal/manifest"
	"unyca-builder/src/internal/version"
)

type Host struct {
	ID           string                 `json:"id"`
	IP           string                 `json:"ip"`
	User         string                 `json:"user"`
	Groups       []string               `json:"groups"`
	SSHKey       string                 `json:"ssh_key,omitempty"`
	InlineKeyB64 string                 `json:"inline_ssh_key_b64,omitempty"`
	Connection   string                 `json:"connection,omitempty"`
	Vars         map[string]interface{} `json:"vars,omitempty"`
}

type Config struct {
	SystemName       string                 `json:"system_name"`
	SystemType       string                 `json:"system_type"`
	BlueprintVersion string                 `json:"blueprint_version,omitempty"`
	BlueprintMeta    map[string]interface{} `json:"blueprint_meta,omitempty"`
	Data             []Host                 `json:"data"`
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func readJSON[T any](path string, v *T) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func repoRoot() string { wd, _ := os.Getwd(); return wd }

func blueprintDir(cfg Config) (string, string, error) {
	root := repoRoot()
	base := filepath.Join(root, "blueprints", cfg.SystemType)
	version := cfg.BlueprintVersion
	if version == "" {
		b, err := ioutil.ReadFile(filepath.Join(base, "LATEST"))
		if err != nil { return "", "", err }
		version = strings.TrimSpace(string(b))
	}
	dir := filepath.Join(base, version)
	if _, err := os.Stat(dir); err != nil { return "", "", fmt.Errorf("blueprint version not found: %s", dir) }
	return dir, version, nil
}

func ensureBuildDir(cfg Config) (string, error) {
	root := repoRoot()
	dir := filepath.Join(root, "builds", cfg.SystemName)
	if err := os.MkdirAll(filepath.Join(dir, "snapshots"), 0755); err != nil { return "", err }
	if err := os.MkdirAll(filepath.Join(dir, "logs"), 0755); err != nil { return "", err }
	return dir, nil
}

func cmdValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	schemaPath := fs.String("schema", "schemas/config.schema.json", "Path to JSON Schema")
	fs.Parse(args)
	if fs.NArg() < 1 { must(errors.New("usage: unyca-builder validate <config.json>")) }
	configPath := fs.Arg(0)
	must(schema.ValidateConfig(configPath, *schemaPath))
	fmt.Println("OK: config is valid")
}

func cmdPlan(args []string) {
	fs := flag.NewFlagSet("plan", flag.ExitOnError)
	schemaPath := fs.String("schema", "schemas/config.schema.json", "Path to JSON Schema")
	fs.Parse(args)
	if fs.NArg() < 1 { must(errors.New("usage: unyca-builder plan <config.json>")) }
	configPath := fs.Arg(0)
	must(schema.ValidateConfig(configPath, *schemaPath))

	var cfg Config
	must(readJSON(configPath, &cfg))

	bpDir, bpVer, err := blueprintDir(cfg)
	must(err)
	must(manifest.Verify(bpDir, version.Version))

	hostCount := len(cfg.Data)
	groupSet := map[string]struct{}{}
	for _, h := range cfg.Data {
		for _, g := range h.Groups { groupSet[g] = struct{}{} }
	}
	var groups []string
	for g := range groupSet { groups = append(groups, g) }
	fmt.Printf("Plan for system=%s type=%s\n", cfg.SystemName, cfg.SystemType)
	fmt.Printf("- Blueprint: %s (version %s)\n", bpDir, bpVer)
	fmt.Printf("- Hosts: %d | Groups: %s\n", hostCount, strings.Join(groups, ","))
	fmt.Println("- Note: Use '--check' in run for Ansible check mode (blueprint permitting).")
}

func copyFile(src, dst string) error {
	in, err := ioutil.ReadFile(src)
	if err != nil { return err }
	return ioutil.WriteFile(dst, in, 0644)
}

func cmdBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	schemaPath := fs.String("schema", "schemas/config.schema.json", "Path to JSON Schema")
	upgrade := fs.Bool("upgrade", False, "Upgrade to version from config or LATEST")
	fs.Parse(args)
	if fs.NArg() < 1 { must(errors.New("usage: unyca-builder build <config.json> [--upgrade]")) }
	configPath := fs.Arg(0)
	must(schema.ValidateConfig(configPath, *schemaPath))

	var cfg Config
	must(readJSON(configPath, &cfg))

	bpDir, versionResolved, err := blueprintDir(cfg)
	must(err)
	must(manifest.Verify(bpDir, version.Version))

	buildDir, err := ensureBuildDir(cfg)
	must(err)

	dst := filepath.Join(buildDir, "data.json")
	must(copyFile(configPath, dst))

	verPath := filepath.Join(buildDir, "blueprint_version.txt")
	if *upgrade {
		must(ioutil.WriteFile(verPath, []byte(versionResolved+"\n"), 0644))
	} else {
		if _, err := os.Stat(verPath); errors.Is(err, fs.ErrNotExist) {
			must(ioutil.WriteFile(verPath, []byte(versionResolved+"\n"), 0644))
		}
	}

	label := time.Now().UTC().Format("20060102-150405Z") + "-build"
	snapDir := filepath.Join(buildDir, "snapshots", label)
	must(os.MkdirAll(snapDir, 0755))
	must(copyFile(dst, filepath.Join(snapDir, "data.json")))
	must(copyFile(verPath, filepath.Join(snapDir, "blueprint_version.txt")))

	must(snapshots.Rotate(buildDir, 100))
	fmt.Println("Build prepared at:", buildDir)
}

func cmdRun(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	tags := fs.String("tags", "", "Comma-separated Ansible tags")
	check := fs.Bool("check", false, "Run Ansible in check mode")
	fs.Parse(args)
	if fs.NArg() < 1 { must(errors.New("usage: unyca-builder run <system_name> [--tags x,y] [--check]")) }
	systemName := fs.Arg(0)

	root := repoRoot()
	buildDir := filepath.Join(root, "builds", systemName)
	dataPath := filepath.Join(buildDir, "data.json")
	if _, err := os.Stat(dataPath); err != nil {
		must(fmt.Errorf("build not found: %s (run 'build' first)", buildDir))
	}
	versionBytes, err := ioutil.ReadFile(filepath.Join(buildDir, "blueprint_version.txt"))
	must(err)
	var cfg Config
	must(readJSON(dataPath, &cfg))
	cfg.BlueprintVersion = strings.TrimSpace(string(versionBytes))

	bpDir, _, err := blueprintDir(cfg)
	must(err)
	must(manifest.Verify(bpDir, version.Version))

	var t []string
	if *tags != "" { t = strings.Split(*tags, ",") }
	extra := map[string]string{}
	if *check { extra["ANSIBLE_CHECK_MODE"] = "1" }

	must(ansible.RunServersYml(ansible.RunOpts{
		BuildDir:      buildDir,
		BlueprintDir:  bpDir,
		BlueprintMeta: cfg.BlueprintMeta,
		Tags:          t,
		ExtraEnv:      extra,
	}))
	fmt.Println("Run completed.")
}

func cmdSnapshot(args []string) {
	fs := flag.NewFlagSet("snapshot", flag.ExitOnError)
	fs.Parse(args)
	if fs.NArg() < 2 { must(errors.New("usage: unyca-builder snapshot <system_name> <label>")) }
	systemName := fs.Arg(0)
	label := fs.Arg(1)

	root := repoRoot()
	buildDir := filepath.Join(root, "builds", systemName)
	dataPath := filepath.Join(buildDir, "data.json")
	versionPath := filepath.Join(buildDir, "blueprint_version.txt")
	if _, err := os.Stat(dataPath); err != nil {
		must(fmt.Errorf("build not found: %s", buildDir))
	}
	ts := time.Now().UTC().Format("20060102-150405Z")
	snapDir := filepath.Join(buildDir, "snapshots", ts+"-"+label)
	must(os.MkdirAll(snapDir, 0755))
	must(copyFile(dataPath, filepath.Join(snapDir, "data.json")))
	must(copyFile(versionPath, filepath.Join(snapDir, "blueprint_version.txt")))

	must(snapshots.Rotate(buildDir, 100))
	fmt.Println("Snapshot created:", snapDir)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: unyca-builder <validate|plan|build|run|snapshot> [args]")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "validate":
		cmdValidate(os.Args[2:])
	case "plan":
		cmdPlan(os.Args[2:])
	case "build":
		cmdBuild(os.Args[2:])
	case "run":
		cmdRun(os.Args[2:])
	case "snapshot":
		cmdSnapshot(os.Args[2:])
	default:
		fmt.Println("unknown command:", os.Args[1])
		os.Exit(2)
	}
}
