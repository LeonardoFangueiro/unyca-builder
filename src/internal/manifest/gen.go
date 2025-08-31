// EN: Generate MANIFEST.json for a blueprint directory.
package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type GenerateOptions struct {
	BlueprintDir string
	MinEngine    string
	MaxEngine    *string
	Excludes     []string // glob patterns over POSIX paths (use "/")
}

type ManifestOut struct {
	Version   string            `json:"version"`
	MinEngine string            `json:"min_engine"`
	MaxEngine *string           `json:"max_engine"`
	Files     map[string]string `json:"files"`
	Signature any               `json:"signature"`
	CreatedAt string            `json:"created_at"`
}

// EN: sha256 of a file as "sha256:<hex>" (unique name to avoid clash with verify.go)
func sha256FileP(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}

func matchesAny(relPOSIX string, patterns []string) bool {
	for _, pat := range patterns {
		if ok, _ := path.Match(pat, relPOSIX); ok {
			return true
		}
	}
	return false
}

// EN: Generate manifest JSON bytes (pretty).
func Generate(opts GenerateOptions) ([]byte, error) {
	bp := opts.BlueprintDir
	info, err := os.Stat(bp)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("blueprint dir not found: %s", bp)
	}

	verBytes, err := os.ReadFile(filepath.Join(bp, "VERSION"))
	if err != nil {
		return nil, fmt.Errorf("VERSION read: %w", err)
	}
	version := strings.TrimSpace(string(verBytes))

	files := map[string]string{}
	err = filepath.WalkDir(bp, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(bp, p)
		relPOSIX := filepath.ToSlash(rel)

		if strings.EqualFold(relPOSIX, "manifest.json") {
			return nil
		}
		if matchesAny(relPOSIX, opts.Excludes) {
			return nil
		}
		sum, err := sha256FileP(p)
		if err != nil {./unyca run game-cp-01 --tags apis,databases
			return fmt.Errorf("hash %s: %w", relPOSIX, err)
		}
		files[relPOSIX] = sum
		return nil
	})
	if err != nil {
		return nil, err
	}

	out := ManifestOut{
		Version:   version,
		MinEngine: opts.MinEngine,
		MaxEngine: opts.MaxEngine,
		Files:     files,
		Signature: nil,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	return json.MarshalIndent(out, "", "  ")
}
