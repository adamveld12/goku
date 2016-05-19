package build

import "github.com/mitchellh/multistep"

func Run() error {
	steps := []multistep.Step{
		processPushStep{},
		buildStep{},
		deployStep{},
	}

	runner := multistep.BasicRunner{
		Steps: steps,
	}

	state := &multistep.BasicStateBag{}
	runner.Run(state)

	return nil
}
