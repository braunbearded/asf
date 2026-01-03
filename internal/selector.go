package internal

import (
	"fmt"
	"os"
	"slices"

	fzf "github.com/junegunn/fzf/src"
)

var (
	FZFDELEMITER       = "|"
	FZFVISUALSEPERATOR = " / "
)

func SelectVaults(channel <-chan []Vault) []Vault {
	var allVaults []Vault
	var selectedNames []string

	inputChan := make(chan string)
	outputChan := make(chan string)

	options, err := fzf.ParseOptions(true, []string{"--multi", "--style", "full", "--delimiter", FZFDELEMITER, "--with-nth", "2.."})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(fzf.ExitError)
	}
	options.Input = inputChan
	options.Output = outputChan

	go func() {
		defer close(inputChan)
		for vaults := range channel {
			allVaults = append(allVaults, vaults...)
			for _, vault := range vaults {
				inputChan <- vault.FormatFZF(FZFDELEMITER, FZFVISUALSEPERATOR)
			}
		}
	}()

	go func() {
		for selection := range outputChan {
			selectedNames = append(selectedNames, selection)
		}
	}()

	_, err = fzf.Run(options)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(fzf.ExitError)
	}

	return FilterVaultsBySelection(allVaults, selectedNames, FZFDELEMITER)
}

func SelectSecrets(channel <-chan Secret) []Secret {
	var allSecrets []Secret
	var fzfSelection []string

	inputChan := make(chan string)
	outputChan := make(chan string)

	options, err := fzf.ParseOptions(true, []string{"--multi", "--style", "full", "--delimiter", FZFDELEMITER, "--with-nth", "2.."})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(fzf.ExitError)
	}
	options.Input = inputChan
	options.Output = outputChan

	go func() {
		defer close(inputChan)
		for secret := range channel {
			allSecrets = append(allSecrets, secret)
			inputChan <- secret.FormatFZF(FZFDELEMITER, FZFVISUALSEPERATOR)
		}
	}()

	go func() {
		for selection := range outputChan {
			fzfSelection = append(fzfSelection, selection)
		}
	}()

	_, err = fzf.Run(options)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(fzf.ExitError)
	}

	return FilterSecretsBySelection(allSecrets, fzfSelection, FZFDELEMITER)
}

func SelectOperation(operationStack []Operation) (*Operation, []Operation) {
	inputChan := make(chan string)
	outputChan := make(chan string)
	var selectedOperation *Operation

	options, err := fzf.ParseOptions(true, []string{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(fzf.ExitError)
	}
	options.Input = inputChan
	options.Output = outputChan

	go func() {
		defer close(inputChan)
		if !slices.Contains(operationStack, ListVersions) && !slices.Contains(operationStack, ListVersionAndGetPasswords) {
			inputChan <- ListVersions.Data().Name
		}
		if (!slices.Contains(operationStack, GetPasswords) || !slices.Contains(operationStack, ListVersions)) && !slices.Contains(operationStack, ListVersionAndGetPasswords) {
			inputChan <- GetPasswords.Data().Name
			inputChan <- ListVersionAndGetPasswords.Data().Name
		}
		inputChan <- EditMetaData.Data().Name
		inputChan <- DeleteSecret.Data().Name
	}()

	go func() {
		for selection := range outputChan {
			operation, errOp := StringToOperation(selection)
			if errOp != nil {
				os.Exit(fzf.ExitError)
			}
			selectedOperation = &operation
		}
	}()

	_, err = fzf.Run(options)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(fzf.ExitError)
	}

	if selectedOperation != nil {
		operationStack = append(operationStack, *selectedOperation)
	}
	return selectedOperation, operationStack
}
