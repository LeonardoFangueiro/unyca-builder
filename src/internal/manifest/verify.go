package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
)

type Manifest struct {
	Version   string             `json:"version"`
	MinEngine string             `json:"min_engine"`
	MaxEngine *string            `json:"max_engine"`
	Files     map[string]string  `json:"files"`
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path); if err != nil { return "", err }
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil { return "", err }
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

func Verify(bpDir string, engineVersion string) error {
	mb, err := os.ReadFile(filepath.Join(bpDir, "MANIFEST.json"))
	if err != nil { return fmt.Errorf("manifest read: %w", err) }
	var m Manifest
	if err := json.Unmarshal(mb, &m); err != nil { return fmt.Errorf("manifest parse: %w", err) }

	ev, err := semver.NewVersion(engineVersion)
	if err != nil { return fmt.Errorf("engine semver: %w", err) }
	minv, err := semver.NewVersion(m.MinEngine)
	if err != nil { return fmt.Errorf("manifest min_engine semver: %w", err) }
	if ev.LessThan(minv) { return fmt.Errorf("builder %s < min_engine %s", ev, minv) }
	if m.MaxEngine != nil {
		maxv, err := semver.NewVersion(*m.MaxEngine)
		if err != nil { return fmt.Errorf("manifest max_engine semver: %w", err) }
		if ev.GreaterThan(maxv) { return fmt.Errorf("builder %s > max_engine %s", ev, maxv) }
	}

	for rel, want := range m.Files {
		p := filepath.Join(bpDir, rel)
		got, err := sha256File(p)
		if err != nil { return fmt.Errorf("hash %s: %w", rel, err) }
		if got != want { return fmt.Errorf("integrity mismatch: %s (got %s want %s)", rel, got, want) }
	}
	return nil
}
