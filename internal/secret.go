package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type Secret struct {
	// Source structs: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets#Secret, https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets#SecretProperties, https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets#SecretAttributes
	ID          string
	ContentType string
	Name        string
	Tags        map[string]string
	Value       string
	Vault       Vault
	Version     string
	Managed     bool
	Client      azsecrets.Client
	Enabled     bool
	Created     time.Time
}

func GetSecrets(vaults []Vault) <-chan Secret {
	secretStream := make(chan Secret)

	go func() {
		defer close(secretStream)
		for _, vault := range vaults {
			client, err := azsecrets.NewClient(vault.VaultURI, vault.Credential, nil)
			if err != nil {
				log.Fatalf("failed to create client for vault %s: %v", vault.ID, err)
			}
			pager := client.NewListSecretPropertiesPager(nil)
			for pager.More() {
				page, err := pager.NextPage(vault.Context)
				if err != nil {
					log.Fatalf("failed to list secrets in vault %s: %v", vault.ID, err)
				}
				for _, secret := range page.Value {
					version := secret.ID.Version()
					if version == "" {
						version = "latest"
					}
					name := secret.ID.Name()

					tags := make(map[string]string)
					for k, v := range secret.Tags {
						if v != nil {
							tags[k] = *v
						}
					}
					secretStream <- Secret{
						ID:      fmt.Sprintf("%s.%s.%s", vault.ID, name, version),
						Name:    name,
						Version: version,
						Client:  *client,
						Vault:   vault,
						Tags:    tags,
						Enabled: *secret.Attributes.Enabled,
						Created: *secret.Attributes.Created,
					}
				}
			}
		}
	}()
	return secretStream
}

func GetVersions(secrets []Secret) <-chan Secret {
	secretStream := make(chan Secret)

	go func() {
		defer close(secretStream)
		for _, secret := range secrets {
			if secret.Version == "latest" {
				secretStream <- secret
			}

			secretClient := secret.Client
			versionsPager := secretClient.NewListSecretPropertiesVersionsPager(secret.Name, nil)

			for versionsPager.More() {
				versionPage, err := versionsPager.NextPage(secret.Vault.Context)
				if err != nil {
					log.Fatalf("failed to list versions for secret %s: %v", secret.Name, err)
				}
				for _, secretVersion := range versionPage.Value {
					name := secretVersion.ID.Name()
					version := secretVersion.ID.Version()

					tags := make(map[string]string)
					for k, v := range secretVersion.Tags {
						if v != nil {
							tags[k] = *v
						}
					}
					secretStream <- Secret{
						ID:      fmt.Sprintf("%s.%s.%s", secret.Vault.ID, name, version),
						Name:    name,
						Version: version,
						Client:  secretClient,
						Vault:   secret.Vault,
						Tags:    tags,
						Enabled: *secretVersion.Attributes.Enabled,
						Created: *secretVersion.Attributes.Created,
					}
				}
			}
		}
	}()
	return secretStream
}

func GetSecretPasswords(secrets []Secret) <-chan Secret {
	secretStream := make(chan Secret)

	go func() {
		defer close(secretStream)
		for _, secret := range secrets {
			version := secret.Version
			if version == "latest" {
				version = ""
			}
			if secret.Value == "" { // todo check for nil
				secretValue, err := secret.Client.GetSecret(secret.Vault.Context, secret.Name, version, nil)
				if err != nil {
					log.Fatalf("failed to get password for secret %s: %v", secret.Name, err)
				}
				secret.Value = *secretValue.Value
			}
			secretStream <- secret
		}
	}()
	return secretStream
}

func GetSecretPasswordsStream(secrets <-chan Secret) <-chan Secret {
	secretStream := make(chan Secret)

	go func() {
		defer close(secretStream)
		for secret := range secrets {
			version := secret.Version
			if version == "latest" {
				version = ""
			}
			if secret.Value == "" { // todo check for nil
				secretValue, err := secret.Client.GetSecret(secret.Vault.Context, secret.Name, version, nil)
				if err != nil {
					log.Fatalf("failed to get password for secret %s: %v", secret.Name, err)
				}
				secret.Value = *secretValue.Value
			}
			secretStream <- secret
		}
	}()
	return secretStream
}

func (secret Secret) FormatFZF(delemiter string, visualSeperator string) string {
	tagsJSON, err := json.Marshal(secret.Tags)
	tagsString := strings.ReplaceAll(string(tagsJSON), `"`, "")
	if err != nil {
		log.Fatalf("Error converting secret tags to string: %v", err)
	}

	password := secret.Value
	if password == "" {
		password = "******"
	}
	created := fmt.Sprintf("{created:%s}", secret.Created.Format("2006-01-02 15:04"))
	enabled := fmt.Sprintf("{enabled:%t}", secret.Enabled)

	return fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s%s%s%s%s", secret.ID, delemiter, secret.Name, visualSeperator, password, visualSeperator, secret.Version, visualSeperator, secret.Vault.Name, visualSeperator, tagsString, visualSeperator, created, visualSeperator, enabled)
}

func FilterSecretsBySelection(secrets []Secret, selections []string, delemiter string) []Secret {
	selectionMap := make(map[string]bool)
	for _, selection := range selections {
		id := strings.Split(selection, delemiter)[0]
		selectionMap[id] = true
	}

	selectedSecrets := make([]Secret, 0, len(selections))
	for _, secret := range secrets {
		if selectionMap[secret.ID] {
			selectedSecrets = append(selectedSecrets, secret)
		}
	}
	return selectedSecrets
}
