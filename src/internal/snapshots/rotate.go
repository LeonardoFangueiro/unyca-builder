package snapshots

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
)

func Rotate(buildDir string, keep int) error {
	root := filepath.Join(buildDir, "snapshots")
	entries, err := ioutil.ReadDir(root)
	if err != nil { return err }
	sort.Slice(entries, func(i, j int) bool { return entries[i].ModTime().After(entries[j].ModTime()) })
	for idx, e := range entries {
		if idx < keep { continue }
		os.RemoveAll(filepath.Join(root, e.Name()))
	}
	return nil
}
