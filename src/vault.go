package asf

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

func FilterVaults(vaults []*armkeyvault.Vault, vaultIDs []string) []*armkeyvault.Vault {
	var result []*armkeyvault.Vault
	for _, vault := range vaults {
		for _, id := range vaultIDs {
			if *vault.ID == id {
				result = append(result, vault)
			}
		}
	}
	return result
}

func GetVaults(subscriptionID string, credentials azcore.TokenCredential, context context.Context) []*armkeyvault.Vault {
	client, err := armkeyvault.NewVaultsClient(subscriptionID, credentials, nil)
	if err != nil {
		log.Fatalf("failed to create vault client: %v", err)
	}

	var vaults []*armkeyvault.Vault

	pager := client.NewListBySubscriptionPager(nil)

	for pager.More() {
		page, err := pager.NextPage(context)
		if err != nil {
			log.Fatalf("failed to get next page: %v", err)
		}

		// https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault#Vault
		vaults = append(vaults, page.Value...)
	}
	return vaults
}

func GetVaults2(context context.Context, credentials azcore.TokenCredential, subscriptionID string) []Vault {
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

func FormatVaultsTable(vaults []*armkeyvault.Vault) string {
	items := make([]interface{}, len(vaults))
	for i, v := range vaults {
		items[i] = v
	}
	columns := []Column{
		{
			Header: "ID", Extractor: func(item interface{}) string { return *item.(*armkeyvault.Vault).ID },
		},
		{
			Header: "Name", Extractor: func(item interface{}) string { return *item.(*armkeyvault.Vault).Name },
		},
		{
			Header: "Tags",
			Extractor: func(item interface{}) string {
				tags := item.(*armkeyvault.Vault).Tags
				jsonString, _ := json.Marshal(tags)
				return string(jsonString)
			},
		},
		{
			Header: "Resource Group", Extractor: func(item interface{}) string {
				vault := item.(*armkeyvault.Vault)
				resourceID, _ := arm.ParseResourceID(*vault.ID)
				return resourceID.ResourceGroupName
			},
		},
		{
			Header: "Location", Extractor: func(item interface{}) string { return *item.(*armkeyvault.Vault).Location },
		},
	}

	return FormatTable(items, columns)
}
