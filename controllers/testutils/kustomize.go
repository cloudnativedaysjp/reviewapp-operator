package testutils

import (
	"bytes"
	"fmt"
	"os/exec"
)

func KustomizeBuildForTest(binPath string) (string, error) {
	cmd := exec.Command(binPath, "build", "../config/test")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(`Error: %v`, stderr.String())
	}
	return stdout.String(), nil
}
