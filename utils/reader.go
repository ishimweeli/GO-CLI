package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ReadInput(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	text = strings.Replace(text, "\n", "", -1)
	return text, nil
}

func ReadInputUntilValid(input *string) error {
	for {
		_, err := fmt.Scanln(input)
		if err != nil {
			fmt.Printf("A valid input is required\n%s:>> ", *input)
		} else {
			break
		}
	}
	return nil
}
