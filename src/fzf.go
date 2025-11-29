package asf

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Sentinel errors for callers to check with errors.Is()
var (
	ErrFzfNotInstalled = errors.New("fzf is not installed or not in PATH")
	ErrUserCancelled   = errors.New("user cancelled selection")
	ErrNoSelection     = errors.New("no selection made")
)

func FzfSelect(input io.Reader, args []string, numFields int, delimiter string) ([]string, error) {
	// Check if fzf is installed
	if _, err := exec.LookPath("fzf"); err != nil {
		return nil, ErrFzfNotInstalled
	}

	cmd := exec.Command("fzf", args...)
	cmd.Stdin = input
	cmd.Stderr = os.Stderr // fzf draws UI to stderr

	// Capture selection from stdout
	outputPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start fzf: %w", err)
	}

	output, err := io.ReadAll(outputPipe)
	if err != nil {
		return nil, fmt.Errorf("failed to read fzf output: %w", err)
	}

	// Check fzf exit code
	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code := exitErr.ExitCode()
			if code == 1 || code == 130 {
				// ESC or Ctrl-C
				return nil, ErrUserCancelled
			}
		}
		return nil, fmt.Errorf("fzf command failed: %w", err)
	}

	// Parse the output
	outputStr := strings.TrimSpace(string(output))
	if outputStr == "" {
		return nil, ErrNoSelection
	}

	lines := strings.Split(outputStr, "\n")
	var results []string

	for _, line := range lines {
		fields := strings.Split(line, delimiter)

		switch {
		case numFields <= 0:
			// Entire line
			results = append(results, strings.TrimSpace(line))

		case numFields == 1:
			// Only first field
			if len(fields) > 0 {
				results = append(results, strings.TrimSpace(fields[0]))
			}

		default:
			// First N fields
			fieldsToReturn := numFields
			if len(fields) < numFields {
				fieldsToReturn = len(fields)
			}

			selected := fields[:fieldsToReturn]
			for i := range selected {
				selected[i] = strings.TrimSpace(selected[i])
			}
			results = append(results, strings.Join(selected, delimiter))
		}
	}

	if len(results) == 0 {
		return nil, ErrNoSelection
	}

	return results, nil
}
