package lib

import (
	"fmt"
	"os/exec"
	"strings"
)

// RunCommand executes a command with some given arguments.
// If the command fails to run, the content of stderr will be part of the returned error.
func RunCommand(cmd string, args ...string) ([]byte, error) {
	out, err := exec.Command(cmd, args...).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			err = fmt.Errorf(string(exitErr.Stderr))
		}

		return nil, fmt.Errorf("Error running command `%s %s`: %v", cmd, strings.Join(args, " "), err)
	}

	return out, nil
}
