package workspace

import (
	"os"
	"path/filepath"
)

// ListMarkdownFiles returns a list of markdown files in the workspace root.
func ListMarkdownFiles(ws *Workspace) ([]string, error) {
	var files []string
	root := ws.Root
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
