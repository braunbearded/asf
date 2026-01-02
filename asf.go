// Package asf handles the interaction with azure and fzf
package asf

import (
	"context"
	"fmt"
	"os"

	"github.com/braunbearded/asf/internal"
	fzf "github.com/junegunn/fzf/src"
)

func Run() {
	// operation on single item (shortcut?)
	// see above +
	// 5. update pw
	// 6. remove

	vaultStream := internal.InitVaults(context.Background())

	selectedVaults := internal.SelectVaults(vaultStream)
	if len(selectedVaults) == 0 {
		fmt.Fprintln(os.Stderr, "No vault selected. Exiting...")
		os.Exit(fzf.ExitError)
	}
	secretStream := internal.GetSecrets(selectedVaults)

	var selectedSecrets []internal.Secret
	var operationStack []internal.Operation

	for {
		selectedSecrets = internal.SelectSecrets(secretStream)
		if len(selectedSecrets) == 0 {
			fmt.Fprintln(os.Stderr, "No secrets selected. Exiting...")
			os.Exit(fzf.ExitError)
		}
		var selectedOperation *internal.Operation
		selectedOperation, operationStack = internal.SelectOperation(operationStack)
		if selectedOperation == nil {
			fmt.Fprintln(os.Stderr, "No operation selected. Exiting...")
			os.Exit(fzf.ExitError)

		}

		switch *selectedOperation {
		case internal.ListVersions:
			secretStream = internal.GetVersions(selectedSecrets)
		case internal.GetPasswords:
			secretStream = internal.GetSecretPasswords(selectedSecrets)
		case internal.ListVersionAndGetPasswords:
			secretStream = internal.GetVersions(selectedSecrets)
			secretStream = internal.GetSecretPasswordsStream(secretStream)
		case internal.EditMetaData:
			fmt.Println("EditMetaData selected")
		// 	edit_meta(*selectedSecrets); confirm(); send_to_azure()
		default:
			fmt.Fprintln(os.Stderr, "Something went wrong. Exiting....")
			os.Exit(fzf.ExitError)
		}
	}
}
