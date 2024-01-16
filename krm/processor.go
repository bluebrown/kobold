package krm

import (
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

type Processor struct{}

type FunctionConfig struct {
	Items []string `json:"items,omitempty"`
}

func (p *Processor) Process(rl *framework.ResourceList) error {
	fc := &FunctionConfig{}
	if err := framework.LoadFunctionConfig(rl.FunctionConfig, fc); err != nil {
		return err
	}
	f := NewImageRefUpdateFilter(nil, fc.Items...)
	return rl.Filter(f)
}
