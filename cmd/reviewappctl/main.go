package main

import (
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/cloudnativedaysjp/reviewapp-operator/cmd/reviewappctl/cmd"
)

func main() {
	cmd.Execute()
}
