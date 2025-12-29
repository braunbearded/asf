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
		var selectedOperation internal.Operation
		selectedOperation, operationStack = internal.SelectOperation(operationStack)

		switch selectedOperation {
		case internal.ListVersions:
			secretStream = internal.GetVersions(selectedSecrets)
		case internal.GetPasswords:
			secretStream = internal.GetSecretPasswords(selectedSecrets)
		case internal.ListVersionAndGetPasswords:
			_ = internal.GetVersions(selectedSecrets)
			// todo collect secrets and save them in selectedSecrets
			secretStream = internal.GetSecretPasswords(selectedSecrets)
			fmt.Println("ListVersionAndGetPasswords password selected")
		case internal.EditMetaData:
			fmt.Println("EditMetaData selected")
		// 	edit_meta(*selectedSecrets); confirm(); send_to_azure()
		default:
			fmt.Println("Something went wrong")
		}
	}
}
