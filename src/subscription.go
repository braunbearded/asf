package asf

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func GetDefaultSubscriptionID() (string, error) {
	cmd := exec.Command("az", "account", "show", "--query", "id", "--output", "tsv")
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to run az account show: %v, stderr: %s", err, stderr.String())
	}
	subID := strings.TrimSpace(out.String())
	if subID == "" {
		return "", fmt.Errorf("got empty subscription id from azure cli")
	}
	return subID, nil
}
