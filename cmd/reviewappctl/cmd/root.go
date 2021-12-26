package cmd

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/validation"
)

// RootCmd is reviewappctl root CLI command.
var RootCmd = &cobra.Command{
	Use:          "reviewappctl",
	SilenceUsage: true,
	Short:        "reviewappctl is tools for management reviewapp-operator",
}

var (
	validator validation.Schema
	builder   *resource.Builder
)

// Execute executes the root command.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	var err error

	configFlags := genericclioptions.NewConfigFlags(true)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(configFlags)
	factory := cmdutil.NewFactory(matchVersionKubeConfigFlags)
	validator, err = factory.Validator(true)
	if err != nil {
		panic(err)
	}
	builder = factory.NewBuilder()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	RootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)
}
