// Command-line interface for unyca-builder
// EN: Minimal CLI with validate, plan, build, run, snapshot.
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"unyca-builder/src/internal/ansible"
	"unyca-builder/src/internal/manifest"
	"unyca-builder/src/internal/schema"
	"unyca-builder/src/internal/snapshots"
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

type stringList []string
func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }


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

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}

func repoRoot() string {
	wd, _ := os.Getwd()
	return wd
}

func blueprintDir(cfg Config) (string, string, error) {
	root := repoRoot()
	base := filepath.Join(root, "blueprints", cfg.SystemType)
	version := cfg.BlueprintVersion
	if version == "" {
		b, err := ioutil.ReadFile(filepath.Join(base, "LATEST"))
		if err != nil {
			return "", "", err
		}
		version = strings.TrimSpace(string(b))
	}
	dir := filepath.Join(base, version)
	if _, err := os.Stat(dir); err != nil {
		return "", "", fmt.Errorf("blueprint version not found: %s", dir)
	}
	return dir, version, nil
}

func ensureBuildDir(cfg Config) (string, error) {
	root := repoRoot()
	dir := filepath.Join(root, "builds", cfg.SystemName)
	if err := os.MkdirAll(filepath.Join(dir, "snapshots"), 0755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Join(dir, "logs"), 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func cmdValidate(args []string) {
	fset := flag.NewFlagSet("validate", flag.ExitOnError)
	schemaPath := fset.String("schema", "schemas/config.schema.json", "Path to JSON Schema")
	fset.Parse(args)
	if fset.NArg() < 1 {
		must(errors.New("usage: unyca-builder validate <config.json>"))
	}
	configPath := fset.Arg(0)
	must(schema.ValidateConfig(configPath, *schemaPath))
	fmt.Println("OK: config is valid")
}

func cmdPlan(args []string) {
	fset := flag.NewFlagSet("plan", flag.ExitOnError)
	schemaPath := fset.String("schema", "schemas/config.schema.json", "Path to JSON Schema")
	fset.Parse(args)
	if fset.NArg() < 1 {
		must(errors.New("usage: unyca-builder plan <config.json>"))
	}
	configPath := fset.Arg(0)

	// Validate first
	must(schema.ValidateConfig(configPath, *schemaPath))

	var cfg Config
	must(readJSON(configPath, &cfg))

	bpDir, bpVer, err := blueprintDir(cfg)
	must(err)
	must(manifest.Verify(bpDir, version.Version))

	// Summary
	hostCount := len(cfg.Data)
	groupSet := map[string]struct{}{}
	for _, h := range cfg.Data {
		for _, g := range h.Groups {
			groupSet[g] = struct{}{}
		}
	}
	var groups []string
	for g := range groupSet {
		groups = append(groups, g)
	}
	fmt.Printf("Plan for system=%s type=%s\n", cfg.SystemName, cfg.SystemType)
	fmt.Printf("- Blueprint: %s (version %s)\n", bpDir, bpVer)
	fmt.Printf("- Hosts: %d | Groups: %s\n", hostCount, strings.Join(groups, ","))
	fmt.Println("- Note: For a check mode, run: unyca-builder run", cfg.SystemName, "--check")
}

func copyFile(src, dst string) error {
	in, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, in, 0644)
}

func cmdBuild(args []string) {
	fset := flag.NewFlagSet("build", flag.ExitOnError)
	schemaPath := fset.String("schema", "schemas/config.schema.json", "Path to JSON Schema")
	upgrade := fset.Bool("upgrade", false, "Upgrade to version from config or LATEST")
	fset.Parse(args)
	if fset.NArg() < 1 {
		must(errors.New("usage: unyca-builder build <config.json> [--upgrade]"))
	}
	configPath := fset.Arg(0)
	must(schema.ValidateConfig(configPath, *schemaPath))

	var cfg Config
	must(readJSON(configPath, &cfg))

	// Resolve blueprint dir/version
	bpDir, versionResolved, err := blueprintDir(cfg)
	must(err)
	must(manifest.Verify(bpDir, version.Version))

	buildDir, err := ensureBuildDir(cfg)
	must(err)

	// Materialize data.json
	dst := filepath.Join(buildDir, "data.json")
	must(copyFile(configPath, dst))

	// Manage blueprint_version.txt
	versionPath := filepath.Join(buildDir, "blueprint_version.txt")
	if *upgrade {
		must(ioutil.WriteFile(versionPath, []byte(versionResolved+"\n"), 0644))
	} else {
		// Keep existing if present; otherwise write resolved
		if _, err := os.Stat(versionPath); errors.Is(err, os.ErrNotExist) {
			must(ioutil.WriteFile(versionPath, []byte(versionResolved+"\n"), 0644))
		}
	}

	// Auto snapshot
	label := time.Now().UTC().Format("20060102-150405Z") + "-build"
	snapDir := filepath.Join(buildDir, "snapshots", label)
	must(os.MkdirAll(snapDir, 0755))
	must(copyFile(dst, filepath.Join(snapDir, "data.json")))
	must(copyFile(versionPath, filepath.Join(snapDir, "blueprint_version.txt")))

	// Rotate snapshots to keep only 100
	must(snapshots.Rotate(buildDir, 100))

	fmt.Println("Build prepared at:", buildDir)
}

func cmdRun(args []string) {
	fset := flag.NewFlagSet("run", flag.ExitOnError)
	tags := fset.String("tags", "", "Comma-separated Ansible tags")
	check := fset.Bool("check", false, "Run Ansible in check mode")
	tee := fset.Bool("tee", true, "Stream Ansible output to console")
    verbosity := fset.Int("v", 0, "Ansible verbosity level (0-4)")
	fset.Parse(args)
	if fset.NArg() < 1 {
		must(errors.New("usage: unyca-builder run <system_name> [--tags x,y] [--check]"))
	}
	systemName := fset.Arg(0)

	// Load build dir
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

	// Resolve blueprint
	bpDir, _, err := blueprintDir(cfg)
	must(err)
	must(manifest.Verify(bpDir, version.Version))

	var t []string
	if *tags != "" {
		t = strings.Split(*tags, ",")
	}

	err = ansible.RunServersYml(ansible.RunOpts{
		BuildDir:      buildDir,
		BlueprintDir:  bpDir,
		BlueprintMeta: cfg.BlueprintMeta,
		Tags:          t,
        ExtraEnv:      map[string]string{},
        Check:         *check,
        Tee:           *tee,
        Verbosity:     *verbosity,

	})
	must(err)
	fmt.Println("Run completed.")
}

func cmdSnapshot(args []string) {
	fset := flag.NewFlagSet("snapshot", flag.ExitOnError)
	fset.Parse(args)
	if fset.NArg() < 2 {
		must(errors.New("usage: unyca-builder snapshot <system_name> <label>"))
	}
	systemName := fset.Arg(0)
	label := fset.Arg(1)

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

func cmdManifest(args []string) {
	fset := flag.NewFlagSet("manifest", flag.ExitOnError)
	bpdir := fset.String("bp", "", "Blueprint dir (e.g., blueprints/<type>/<ver>)")
	minEngine := fset.String("min-engine", "0.1.0", "Minimum engine version")
	maxEngine := fset.String("max-engine", "none", "Maximum engine version or 'none'")
	write := fset.Bool("write", false, "Write to <bp>/MANIFEST.json instead of stdout")
	var excludes stringList
	fset.Var(&excludes, "exclude", "Glob to exclude (repeatable)")
	fset.Parse(args)

	if *bpdir == "" {
		must(errors.New("usage: unyca-builder manifest --bp <dir> [--min-engine X] [--max-engine Y|none] [--exclude pat] [--write]"))
	}

	var maxPtr *string
	if strings.ToLower(*maxEngine) != "none" && strings.TrimSpace(*maxEngine) != "" {
		m := *maxEngine
		maxPtr = &m
	}

	jsonb, err := manifest.Generate(manifest.GenerateOptions{
		BlueprintDir: *bpdir,
		MinEngine:    *minEngine,
		MaxEngine:    maxPtr,
		Excludes:     excludes,
	})
	must(err)

	if *write {
		dst := filepath.Join(*bpdir, "MANIFEST.json")
		must(ioutil.WriteFile(dst, jsonb, 0644))
		fmt.Println("Wrote", dst)
		return
	}
	os.Stdout.Write(jsonb)
	fmt.Println()
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
	case "manifest":
		cmdManifest(os.Args[2:])
	default:
		fmt.Println("unknown command:", os.Args[1])
		os.Exit(2)
	}
}
