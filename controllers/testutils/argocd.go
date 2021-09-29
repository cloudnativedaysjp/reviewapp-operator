package testutils

import (
	"bytes"
	"fmt"
	"os/exec"
)

const (
	argocdPassword     = "admin-password"
	argocdAppNamespace = "argocd"
)

func SyncArgoCDApplication(binPath string, app string) error {
	cmd := exec.Command(binPath, "login", "--username", "admin", "--password", argocdPassword, "--port-forward", "--port-forward-namespace", argocdAppNamespace)
	stderr := bytes.Buffer{}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(`Error: %v`, stderr.String())
	}

	cmd = exec.Command(binPath, "app", "sync", app, "--port-forward", "--port-forward-namespace", argocdAppNamespace)
	stderr = bytes.Buffer{}
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(`Error: %v`, stderr.String())
	}

	return nil
}
