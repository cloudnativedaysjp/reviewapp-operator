package testutils

import (
	"bytes"
	"fmt"
	"os/exec"
)

func KustomizeBuildForTest() (string, error) {
	cmd := exec.Command("kustomize", "build", "../config/test")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(`Error: %v`, stderr.String())
	}
	return stdout.String(), nil
}
