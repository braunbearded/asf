// Package asf handels the interaction with azure and fzf
package asf

import (
	"context"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	fzf "github.com/junegunn/fzf/src"
)

type Operation int

const (
	ListVersions Operation = iota
	GetPasswords
	ListVersionAndGetPasswords
	EditMetaData
)

type OperationData struct {
	Name        string
	Description string
	Delemiter   string
}

var allOperations = map[Operation]OperationData{
	ListVersions:               {"list-versions", "List versions for selected items", ";"},
	GetPasswords:               {"get-passwords", "Get passwords for selected items", ";"},
	ListVersionAndGetPasswords: {"list-version-get-password", "List versions and get passwords for selected items", ";"},
	EditMetaData:               {"edit-meta", "Edit meta data for selected items in $EDITOR", ";"},
}

func (operation Operation) Data() OperationData {
	return allOperations[operation]
}

func StringToOperation(name string) (Operation, bool) {
	for item, data := range allOperations {
		if data.Name == name {
			return item, true
		}
	}
	return 0, false
}

type Vault struct {
	// Source struct https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault#Vault
	ID             string
	Name           string
	Tags           map[string]string
	Location       string
	Context        context.Context
	Credential     azcore.TokenCredential
	SubscriptionID string
	TenantID       string
	VaultURI       string
	// ResourceGroup string //TODO check if really needed
}

type Secret struct {
	// Source structs: https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets#Secret, https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets#SecretProperties, https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets#SecretAttributes
	ContentType string
	Name        string
	Tags        map[string]string
	Value       string
	Vault       *Vault
	Version     string
	Managed     bool
	Client      azsecrets.Client
	// Attributes // TODO check if needed
}

func initVaults(context context.Context) <-chan []Vault {
	channel := make(chan []Vault)

	go func() {
		defer close(channel)

		credential, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			fmt.Errorf("failed to get credential: %v", err)
			return
		}

		subscriptionID, err := GetDefaultSubscriptionID()
		if err != nil {
			fmt.Errorf("failed to get subscription: %v", err)
			return
		}

		vaults := GetVaults2(context, credential, subscriptionID)
		channel <- vaults
	}()

	return channel
}

func selectVaults(channel <-chan []Vault) []Vault {
	var allVaults []Vault
	var selectedNames []string

	inputChan := make(chan string)
	outputChan := make(chan string)

	options, err := fzf.ParseOptions(true, []string{})
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(fzf.ExitError)
	}
	options.Input = inputChan
	options.Output = outputChan

	// Feed vault names to fzf as they arrive
	go func() {
		defer close(inputChan)
		for vaults := range channel {
			allVaults = append(allVaults, vaults...)
			for _, vault := range vaults {
				inputChan <- vault.Name // TODO improve formatting
			}
		}
	}()

	// Read selections in go routine because it will deadlock otherwise
	go func() {
		defer close(outputChan)
		for selection := range outputChan {
			selectedNames = append(selectedNames, selection)
		}
	}()

	_, err = fzf.Run(options)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(fzf.ExitError)
	}

	// Return matching vaults
	var selectedVaults []Vault
	for _, vault := range allVaults {
		if slices.Contains(selectedNames, vault.Name) {
			selectedVaults = append(selectedVaults, vault)
		}
	}

	return selectedVaults
}

func getSecrets(vaults []Vault) <-chan Secret {
	secretStream := make(chan Secret)

	go func() {
		defer close(secretStream)
		for _, vault := range vaults {
			client, err := azsecrets.NewClient(vault.VaultURI, vault.Credential, nil)
			if err != nil {
				log.Fatalf("failed to create client for vault %s: %w", vault.ID, err)
			}
			pager := client.NewListSecretPropertiesPager(nil)
			for pager.More() {
				page, err := pager.NextPage(vault.Context)
				if err != nil {
					log.Fatalf("failed to list secrets in vault %s: %w", vault.ID, err)
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
					}
				}
			}
		}
	}()
	return secretStream
}

func formatSecretForFzf(secret Secret) string {
	password := secret.Value
	if password == "" {
		password = "******"
	}
	return fmt.Sprintf("%s | %s | %s", secret.Name, secret.Version, password)
}

func selectSecrets(channel <-chan Secret) []Secret {
	var allSecrets []Secret
	var selectedSecrets []Secret
	var fzfSelection []string

	inputChan := make(chan string)
	outputChan := make(chan string)

	options, err := fzf.ParseOptions(true, []string{"--multi"})
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
			inputChan <- formatSecretForFzf(secret) // TODO improve formatting
		}
	}()

	go func() {
		defer close(outputChan)
		for selection := range outputChan {
			fzfSelection = append(fzfSelection, selection)
		}
	}()

	_, err = fzf.Run(options)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(fzf.ExitError)
	}

	for _, secret := range allSecrets {
		if slices.Contains(fzfSelection, formatSecretForFzf(secret)) {
			selectedSecrets = append(selectedSecrets, secret)
		}
	}

	return selectedSecrets
}

func selectOperation(operationStack *[]Operation) Operation { // todo return slice and operation instead of using pointer
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
		if !slices.Contains(*operationStack, ListVersions) {
			inputChan <- ListVersions.Data().Name
		}
		if !slices.Contains(*operationStack, GetPasswords) || !slices.Contains(*operationStack, ListVersions) {
			inputChan <- GetPasswords.Data().Name
		}
		inputChan <- ListVersionAndGetPasswords.Data().Name
	}()

	go func() {
		defer close(outputChan)
		for selection := range outputChan {
			operation, err := StringToOperation(selection)
			if !err {
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

	if selectedOperation == nil {
		fmt.Println("Some weird happen in selectOperation")
		os.Exit(fzf.ExitError)
	}
	*operationStack = append(*operationStack, *selectedOperation)
	return *selectedOperation
}

func getVersions(secrets []Secret) <-chan Secret {
	secretStream := make(chan Secret)

	go func() {
		defer close(secretStream)
		for _, secret := range secrets {
			secretStream <- secret

			secretClient := secret.Client
			versionsPager := secretClient.NewListSecretPropertiesVersionsPager(secret.Name, nil)

			for versionsPager.More() {
				versionPage, err := versionsPager.NextPage(context.Background()) // todo dirty fix for context -> null pointer needs to be fixed
				if err != nil {
					log.Fatalf("failed to list versions for secret %s: %w", secret.Name, err)
				}
				for _, secretVersion := range versionPage.Value {
					secretStream <- Secret{Name: secretVersion.ID.Name(), Version: secretVersion.ID.Version(), Client: secretClient}
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
				secretValue, err := secret.Client.GetSecret(context.Background(), secret.Name, version, nil)
				if err != nil {
					log.Fatalf("failed to get password for secret %s: %w", secret.Name, err)
				}
				secret.Value = *secretValue.Value
			}
			secretStream <- secret
		}
	}()
	return secretStream
}

func Run() {
	// operation on single item (shortcut?)
	// see above +
	// 5. update pw
	// 6. remove

	vaultStream := initVaults(context.Background())

	selectedVaults := selectVaults(vaultStream)
	if len(selectedVaults) == 0 {
		fmt.Fprintln(os.Stderr, "No vault selected. Exiting...")
		os.Exit(fzf.ExitError)
	}
	secretStream := getSecrets(selectedVaults)

	var selectedSecrets []Secret
	var operationStack []Operation

	for {
		selectedSecrets = selectSecrets(secretStream)
		if len(selectedSecrets) == 0 {
			fmt.Fprintln(os.Stderr, "No secrets selected. Exiting...")
			os.Exit(fzf.ExitError)
		}
		selectedOperation := selectOperation(&operationStack)
		switch selectedOperation {
		case ListVersions:
			secretStream = getVersions(selectedSecrets)
		case GetPasswords:
			secretStream = GetSecretPasswords(selectedSecrets)
		case ListVersionAndGetPasswords:
			_ = getVersions(selectedSecrets)
			// todo collect secrets and save them in selectedSecrets
			secretStream = GetSecretPasswords(selectedSecrets)
			fmt.Println("ListVersionAndGetPasswords password selected")
		case EditMetaData:
			fmt.Println("EditMetaData selected")
		// 	edit_meta(*selectedSecrets); confirm(); send_to_azure()
		default:
			fmt.Println("Something went wrong")
		}
	}
}
