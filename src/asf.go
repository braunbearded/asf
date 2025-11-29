// Package asf handels the interaction with azure and fzf
package asf

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

func Run() {
	ctx := context.Background()

	// Use Azure CLI token / DefaultAzureCredential
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to get credential: %v", err)
	}

	subscriptionID, err := GetDefaultSubscriptionID()
	if err != nil {
		log.Fatalf("failed to get subscription: %v", err)
	}
	fmt.Println("Using subscription:", subscriptionID)

	vaults := GetVaults(subscriptionID, cred, ctx)

	vaultsTable := FormatVaultsTable(vaults)

	// Build fzf command
	selectVaultArgs := []string{
		"--header-lines=1", // Skip the header and separator lines
		"--delimiter=\t",   // Tab delimiter (literal tab character in Go string)
		"--with-nth=2..",   // Show all fields
		"--multi",
	}

	selectedVaults, err := FzfSelect(strings.NewReader(vaultsTable), selectVaultArgs, 1, "\t")
	if err != nil {
		if errors.Is(err, ErrUserCancelled) || errors.Is(err, ErrNoSelection) {
			os.Exit(0)
		}

		// Any other error is real and should be reported
		log.Fatalf("Error: %v", err)
	}

	selectedOperationArgs := []string{
		"--delimiter=\t", // Tab delimiter (literal tab character in Go string)
		"--with-nth=2..", // Show all fields
	}
	selectedOperation, err := FzfSelect(strings.NewReader("list\tlist\nadd\tadd"), selectedOperationArgs, 1, "\t")
	if err != nil {
		if errors.Is(err, ErrUserCancelled) || errors.Is(err, ErrNoSelection) {
			os.Exit(0)
		}
		log.Fatalf("Error: %v", err)
	}

	var versions []*azsecrets.SecretProperties

	if selectedOperation[0] == "list" {
		for _, selectedVaultId := range selectedVaults {
			for _, vault := range vaults {
				if *vault.ID == selectedVaultId {
					fmt.Println("Match")
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
		selectedSecrets, err := FzfSelect(strings.NewReader(secretsTable), selectSecretsArgs, 2, "\t")
		if err != nil {
			if errors.Is(err, ErrUserCancelled) || errors.Is(err, ErrNoSelection) {
				os.Exit(0)
			}
			log.Fatalf("Error: %v", err)
		}

		if len(selectedSecrets) == 1 {
			selectedKeyOperation, err := FzfSelect(strings.NewReader("remove\tremove\nshow-pw\tshow passwod\nupdate-meta\tupdate metadata\nupdate-pw\tupdate password\nnew-version\tadd new version"), selectedOperationArgs, 1, "\t")
			if err != nil {
				if errors.Is(err, ErrUserCancelled) || errors.Is(err, ErrNoSelection) {
					os.Exit(0)
				}
				log.Fatalf("Error: %v", err)
			}
			fmt.Println(selectedKeyOperation)
		}

		if len(selectedSecrets) > 1 {
			selectedKeyOperation, err := FzfSelect(strings.NewReader("show-pw\tshow passwod\nupdate-meta\tupdate metadata"), selectedOperationArgs, 1, "\t")
			if err != nil {
				if errors.Is(err, ErrUserCancelled) || errors.Is(err, ErrNoSelection) {
					os.Exit(0)
				}
				log.Fatalf("Error: %v", err)
			}
			fmt.Println(selectedKeyOperation)
		}
	}
}
