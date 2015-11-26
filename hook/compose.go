package hook

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Compose(proj repository) error {
	fmt.Println("Docker compose detected... @", proj.TargetFilePath)
	dockCompose := exec.Command("/bin/docker-compose", "up", "--force-recreate")

	abs, _ := filepath.Abs(filepath.Dir(proj.TargetFilePath))

	dockCompose.Dir = abs
	dockCompose.Stdout = os.Stdout
	dockCompose.Stderr = os.Stderr

	if err := dockCompose.Run(); err != nil {
		return errors.New(fmt.Sprintln("error running docker-compose\n", err.Error()))
	}

	return errors.New("Not supported")
}
