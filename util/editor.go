package util

import (
	"os"
	"os/exec"

	e "github.com/pkg/errors"
)

func FindEditor(name string) (string, error) {
	if name == "" {
		name = os.Getenv("VISUAL")
	}
	if name == "" {
		name = os.Getenv("EDITOR")
	}
	if name == "" {
		name = "vi"
	}
	path, err := exec.LookPath(name)
	if err != nil {
		return "", e.Wrapf(err, "find editor '%s'", name)
	}
	return path, nil
}

func EditFile(filePath string, editorPath string) error {
	cmd := exec.Command(editorPath, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
