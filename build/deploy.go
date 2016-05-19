package build

import "github.com/mitchellh/multistep"

type deployStep struct{}

func (ds deployStep) Run(state multistep.StateBag) multistep.StepAction {
	return multistep.ActionContinue
}

func (ds deployStep) Cleanup(state multistep.StateBag) {}
