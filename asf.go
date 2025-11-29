package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"encoding/json"
)

func getDefaultSubscriptionID() (string, error) {
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

func tagsToJson(tags map[string]*string) string {
	tagsJSON := "{}"
	if tags != nil && len(tags) > 0 {
		tagsBytes, err := json.Marshal(tags)
		if err == nil {
			tagsJSON = string(tagsBytes)
		}
	}
	return tagsJSON
}

func main() {

	// Use Azure CLI token / DefaultAzureCredential
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to get credential: %v", err)
	}

	subscriptionID, err := getDefaultSubscriptionID()
	if err != nil {
		log.Fatalf("failed to get subscription: %v", err)
	}
	fmt.Println("Using subscription:", subscriptionID)

	client, err := armkeyvault.NewVaultsClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create vault client: %v", err)
	}

	ctx := context.Background()

	var vaults []*armkeyvault.Vault

	pager := client.NewListBySubscriptionPager(nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)

		if err != nil {
			log.Fatalf("failed to get next page: %v", err)
		}

		// https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault#Vault
		vaults = append(vaults, page.Value...)
	}

	vaultsTable := FormatVaultsTable(vaults)

	// Build fzf command
	selectVaultArgs := []string{
		"--header-lines=1", // Skip the header and separator lines
		"--delimiter=\t",   // Tab delimiter (literal tab character in Go string)
		"--with-nth=2..",   // Show all fields
		"--multi",
	}

	selectedVaults, err := FzfSelect(strings.NewReader(vaultsTable), selectVaultArgs, 1, "\t")

	fmt.Println(selectedVaults)

	selectedOperationArgs := []string{
		"--delimiter=\t", // Tab delimiter (literal tab character in Go string)
		"--with-nth=2..", // Show all fields
	}
	selectedOperation, err := FzfSelect(strings.NewReader("list\tlist\nadd\tadd"), selectedOperationArgs, 1, "\t")

	fmt.Println(selectedOperation)

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
			log.Fatalf("error happend: %w", err)
		}

		fmt.Println(selectedSecrets)
		fmt.Printf("Len: %s", len(selectedSecrets))

		if len(selectedSecrets) == 1 {
			selectedKeyOperation, err := FzfSelect(strings.NewReader("remove\tremove\nshow-pw\tshow passwod\nupdate-meta\tupdate metadata\nupdate-pw\tupdate password\nnew-version\tadd new version"), selectedOperationArgs, 1, "\t")

			if err != nil {
				log.Fatalf("error happend: %w", err)
			}
			fmt.Println(selectedKeyOperation)
		}

		if len(selectedSecrets) > 1 {
			selectedKeyOperation, err := FzfSelect(strings.NewReader("show-pw\tshow passwod\nupdate-meta\tupdate metadata"), selectedOperationArgs, 1, "\t")

			if err != nil {
				log.Fatalf("error happend: %w", err)
			}
			fmt.Println(selectedKeyOperation)
		}
	}
}

