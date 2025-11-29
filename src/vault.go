package asf

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"
)

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
				return tagsToJson(tags)
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
