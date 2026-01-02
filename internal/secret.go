package internal

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
)

type Secret struct {
	// Source structs: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets#Secret, https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets#SecretProperties, https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets#SecretAttributes
	ContentType string
	Name        string
	Tags        map[string]string
	Value       string
	Vault       Vault
	Version     string
	Managed     bool
	Client      azsecrets.Client
	// Attributes // TODO check if needed
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
					secretStream <- Secret{
						Name:    secret.ID.Name(),
						Version: version,
						Client:  *client,
						Vault:   vault,
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
			secretStream <- secret

			secretClient := secret.Client
			versionsPager := secretClient.NewListSecretPropertiesVersionsPager(secret.Name, nil)

			for versionsPager.More() {
				versionPage, err := versionsPager.NextPage(secret.Vault.Context)
				if err != nil {
					log.Fatalf("failed to list versions for secret %s: %v", secret.Name, err)
				}
				for _, secretVersion := range versionPage.Value {
					secretStream <- Secret{Name: secretVersion.ID.Name(), Version: secretVersion.ID.Version(), Client: secretClient, Vault: secret.Vault}
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
