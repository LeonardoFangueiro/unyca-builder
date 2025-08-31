package ansible

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RunOpts struct {
	BuildDir      string
	BlueprintDir  string
	BlueprintMeta map[string]any
	Tags          []string
	ExtraEnv      map[string]string
}

func timestamp() string { return time.Now().UTC().Format("20060102-150405Z") }

func findEntrypoint(dir string) (string, error) {
	candidates := []string{"orchestrator.yml","orchestrator.yaml","orchestrator.json","servers.yml","servers.yaml","servers.json"}
	for _, c := range candidates {
		p := filepath.Join(dir, c)
		if st, err := os.Stat(p); err == nil && !st.IsDir() { return p, nil }
	}
	return "", errors.New("no orchestrator|servers playbook found in " + dir)
}

func RunServersYml(o RunOpts) error {
	metaJSON, _ := json.Marshal(o.BlueprintMeta)
	entry, err := findEntrypoint(o.BlueprintDir)
	if err != nil { return err }
	args := []string{entry, "-i", "localhost,", "-c", "local", "-e", "@" + filepath.Join(o.BuildDir, "data.json"), "-e", "blueprint_meta=" + string(metaJSON)}
	if len(o.Tags) > 0 { args = append(args, "--tags", strings.Join(o.Tags, ",")) }
	cmd := exec.Command("ansible-playbook", args...)
	env := os.Environ()
	env = append(env, "UNYCA_BUILD_DIR="+o.BuildDir)
	env = append(env, "ANSIBLE_CONFIG="+filepath.Join(o.BlueprintDir, "ansible.cfg"))
	for k,v := range o.ExtraEnv { env = append(env, k+"="+v) }
	cmd.Env = env
	_ = os.MkdirAll(filepath.Join(o.BuildDir, "logs"), 0755)
	logFile := filepath.Join(o.BuildDir, "logs", "ansible-"+timestamp()+".log")
	lf, _ := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer lf.Close()
	cmd.Stdout = lf
	cmd.Stderr = lf
	return cmd.Run()
}
