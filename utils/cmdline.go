package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// AskForConfirmation asks the user if he wants to continue.
func AskForConfirmation(s string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true, nil
		} else if response == "n" || response == "no" {
			return false, nil
		}
	}
}

// PromptChoice gets input from the user
func PromptChoice(question string, defaultValue string, allowEmpty, nonInteractive bool) (string, error) {
	// Is it non-interactive?
	if nonInteractive {
		switch {
		case defaultValue != "":
			return defaultValue, nil
		case allowEmpty:
			return "", nil
		default:
			return "", fmt.Errorf("%s\nRequested non-interactive command, but a necessary parameter was not provided.", question)
		}
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(question)
		if defaultValue != "" {
			fmt.Printf(" [%s]", defaultValue)
		}
		fmt.Printf(" ")

		response, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}

		response = strings.TrimSpace(response)

		if response == "" && defaultValue == "" && !allowEmpty {
			continue
		}

		if defaultValue != "" && response == "" {
			response = defaultValue
		}

		return response, nil
	}
}
