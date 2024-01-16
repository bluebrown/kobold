package main

import (
	"fmt"
	"os"

	"github.com/bluebrown/kobold/krm"
	"sigs.k8s.io/kustomize/kyaml/fn/framework/command"
)

func main() {
	p := &krm.Processor{}
	cmd := command.Build(p, command.StandaloneEnabled, false)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "%v\n", err)
		os.Exit(1)
	}
}
