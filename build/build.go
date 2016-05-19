package build

import (
	"fmt"

	"github.com/mitchellh/multistep"
)

type buildStep struct{}

func (bs buildStep) Run(state multistep.StateBag) multistep.StepAction {
	repoName := state.Get("repoName").(string)
	repoPath := state.Get("repoPath").(string)
	branch := state.Get("branch").(string)
	project := state.Get("project").(Project)

	fmt.Println("repo: %s, path: %s, branch: %s, project: %s", repoName, repoPath, branch, project.TargetFilePath)

	containerImage := fmt.Sprintf("%s-%s", branch)

	state.Put("containerImage", containerImage)

	return multistep.ActionContinue
}

func (bs buildStep) Cleanup(state multistep.StateBag) {}
