package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func FzfSelect(input io.Reader, args []string, numFields int, delimiter string) ([]string, error) {
	// Check if fzf is installed
	if _, err := exec.LookPath("fzf"); err != nil {
		return nil, errors.New("fzf is not installed or not in PATH")
	}
	cmd := exec.Command("fzf", args...)
	// Connect to terminal for interactive UI
	cmd.Stdin = input
	cmd.Stderr = os.Stderr // fzf writes its UI to stderr
	// Create a pipe to capture the selection output
	outputPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	// Start fzf
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start fzf: %w", err)
	}
	// Read the output
	output, err := io.ReadAll(outputPipe)
	if err != nil {
		return nil, fmt.Errorf("failed to read fzf output: %w", err)
	}
	// Wait for fzf to finish
	if err := cmd.Wait(); err != nil {
		// fzf returns exit code 1 if user cancels (ESC) or exit code 130 for Ctrl-C
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 || exitErr.ExitCode() == 130 {
				return nil, errors.New("user cancelled selection")
			}
		}
		return nil, fmt.Errorf("fzf command failed: %w", err)
	}
	// Parse the output - extract specified fields from each selected line
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return nil, errors.New("no selection made")
	}
	lines := strings.Split(outputStr, "\n")
	var results []string
	for _, line := range lines {
		// Split by the specified delimiter
		fields := strings.Split(line, delimiter)
		
		// Handle different numFields values
		if numFields <= 0 {
			// Return the entire line if numFields is 0 or negative
			results = append(results, strings.TrimSpace(line))
		} else if numFields == 1 {
			// Return only the first field
			if len(fields) > 0 {
				results = append(results, strings.TrimSpace(fields[0]))
			}
		} else {
			// Return the specified number of fields joined back together
			fieldsToReturn := numFields
			if len(fields) < numFields {
				fieldsToReturn = len(fields)
			}
			selectedFields := fields[:fieldsToReturn]
			// Trim each field and join them back with the delimiter
			for i := range selectedFields {
				selectedFields[i] = strings.TrimSpace(selectedFields[i])
			}
			results = append(results, strings.Join(selectedFields, delimiter))
		}
	}
	if len(results) == 0 {
		return nil, errors.New("no valid selection")
	}
	return results, nil
}

