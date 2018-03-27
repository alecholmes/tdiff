package lib

import (
	"fmt"
	"os/exec"
)

func RunCommand(cmd string, args ...string) ([]byte, error) {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf(string(exitErr.Stderr))
		}
		return nil, err
	}

	return out, nil
}
