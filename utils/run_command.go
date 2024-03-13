package utils

import (
	"errors"
	"fmt"
	"os/exec"
)

func RunCommand(command string) (string, error) {
	if command == "" {
		return "", errors.New("no command provided")
	}

	args := []string{"-l"}

	cmd := exec.Command(command, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error running command: %v\n", err)
		return "", err
	}

	fmt.Printf("Command Output:\n%s\n", string(output))

	return "Command run successfully", nil
}
