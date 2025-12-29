package internal

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

func GetVaults(context context.Context, credentials azcore.TokenCredential, subscriptionID string) []Vault {
	client, err := armkeyvault.NewVaultsClient(subscriptionID, credentials, nil)
	if err != nil {
		log.Fatalf("failed to create vault client: %v", err)
	}

	var vaults []Vault

	pager := client.NewListBySubscriptionPager(nil)

	for pager.More() {
		page, err := pager.NextPage(context)
		if err != nil {
			log.Fatalf("failed to get next page: %v", err)
		}

		// https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault#Vault

		for _, vault := range page.Value {
			tags := make(map[string]string)
			for k, v := range vault.Tags {
				if v != nil {
					tags[k] = *v
				}
			}
			vaults = append(vaults, Vault{ID: *vault.ID, Name: *vault.Name, Tags: tags, Location: *vault.Location, Context: context, Credential: credentials, SubscriptionID: subscriptionID, TenantID: *vault.Properties.TenantID, VaultURI: *vault.Properties.VaultURI})
		}
	}
	return vaults
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

func InitVaults(context context.Context) <-chan []Vault {
	channel := make(chan []Vault)

	go func() {
		defer close(channel)

		credential, err := azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			log.Fatalf("failed to get credential: %v", err)
		}

		subscriptionID, err := GetDefaultSubscriptionID()
		if err != nil {
			log.Fatalf("failed to get subscription: %v", err)
		}

		vaults := GetVaults(context, credential, subscriptionID)
		channel <- vaults
	}()

	return channel
}
