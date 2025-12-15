// Package asf handels the interaction with azure and fzf
package asf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

func FzfSelectOrExit(input io.Reader, fzfArgs []string, numFields int, delemiter string) []string {
	result, err := FzfSelect(input, fzfArgs, numFields, delemiter)
	if err != nil {
		if errors.Is(err, ErrUserCancelled) || errors.Is(err, ErrNoSelection) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return result
}

func Run() {
	ctx := context.Background()

	pr, pw := io.Pipe()

	errorChannel := make(chan error, 1)
	credChannel := make(chan *azidentity.DefaultAzureCredential, 1)
	vaultChannel := make(chan []*armkeyvault.Vault, 1)
	subscriptionChannel := make(chan string, 1)

	go func() {
		defer pw.Close()
		// Use Azure CLI token / DefaultAzureCredential
		cred, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			errorChannel <- fmt.Errorf("failed to get credential: %v", err)
			return
		}
		credChannel <- cred

		subscriptionID, err := GetDefaultSubscriptionID()
		if err != nil {
			errorChannel <- fmt.Errorf("failed to get subscription: %v", err)
			return
		}
		subscriptionChannel <- subscriptionID
		// fmt.Println("Using subscription:", subscriptionID)

		vaults := GetVaults(subscriptionID, cred, ctx)
		vaultChannel <- vaults

		vaultsTable := FormatVaultsTable(vaults)
		fmt.Fprintf(pw, "%s\n", vaultsTable)
	}()

	// Build fzf command
	selectVaultArgs := []string{
		"--header-lines=1", // Skip the header and separator lines
		"--delimiter=\t",   // Tab delimiter (literal tab character in Go string)
		"--with-nth=2..",   // Show all fields
		"--multi",
		"--prompt=Select vault> ",
		"--header=TAB: Select\nENTER: Continue\n\n",
		"--layout=reverse",
		"--info=inline",
		"--border",
		// "--style=full", // todo required newer fzf version
		// "--input-border", // todo required newer fzf version
		// "--info=inline-right", // todo required newer fzf version
		// "--header-border", // todo required newer fzf version
		// "--gap" //todo check if usefull
	}

	selectedVaultIDs := FzfSelectOrExit(pr, selectVaultArgs, 1, "\t")

	cred := <-credChannel
	// subscriptionID := <-subscriptionChannel
	vaults := <-vaultChannel

	selectedVaults := FilterVaults(vaults, selectedVaultIDs)

	selectedOperationArgs := []string{
		"--delimiter=\t", // Tab delimiter (literal tab character in Go string)
		"--with-nth=2..", // Show all fields
	}
	selectedOperation := FzfSelectOrExit(strings.NewReader("list\tlist\nadd\tadd"), selectedOperationArgs, 1, "\t")

	var versions []*azsecrets.SecretProperties

	if selectedOperation[0] == "list" {
		for _, vault := range selectedVaults {
			secretClient, err := azsecrets.NewClient(*vault.Properties.VaultURI, cred, nil)
			if err != nil {
				log.Fatalf("failed to create client for vault %s: %w", *vault.ID, err)
			}

			secretsPager := secretClient.NewListSecretPropertiesPager(nil)
			for secretsPager.More() {
				page, err := secretsPager.NextPage(ctx)
				if err != nil {
					log.Fatalf("failed to list secrets in vault %s: %w", *vault.ID, err)
				}

				for _, secretItem := range page.Value {
					secretName := secretItem.ID.Name()

					versionsPager := secretClient.NewListSecretPropertiesVersionsPager(secretName, nil)
					for versionsPager.More() {
						versionPage, err := versionsPager.NextPage(ctx)
						if err != nil {
							log.Fatalf("failed to list versions for secret %s: %w", secretName, err)
						}
						for _, versionItem := range versionPage.Value {
							versions = append(versions, versionItem)
						}
					}
				}
			}
		}
		secretsTable := FormatSecretsTable(versions)

		selectSecretsArgs := []string{
			"--header-lines=1", // Skip the header and separator lines
			"--delimiter=\t",   // Tab delimiter (literal tab character in Go string)
			"--with-nth=3..",   // Show all fields
			"--multi",
			"--preview=/usr/bin/az keyvault secret show --id '{1}'",
		}
		selectedSecrets := FzfSelectOrExit(strings.NewReader(secretsTable), selectSecretsArgs, 2, "\t")

		if len(selectedSecrets) == 1 {
			selectedKeyOperation := FzfSelectOrExit(strings.NewReader("remove\tremove\nshow-pw\tshow passwod\nupdate-meta\tupdate metadata\nupdate-pw\tupdate password\nnew-version\tadd new version"), selectedOperationArgs, 1, "\t")
			fmt.Println(selectedKeyOperation)
		}

		if len(selectedSecrets) > 1 {
			selectedKeyOperation := FzfSelectOrExit(strings.NewReader("show-pw\tshow passwod\nupdate-meta\tupdate metadata"), selectedOperationArgs, 1, "\t")
			fmt.Println(selectedKeyOperation)
		}
	}
}
