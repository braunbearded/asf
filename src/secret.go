package asf

import (
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"fmt"
)

func FormatSecretsTable(secrets []*azsecrets.SecretProperties) string {
	items := make([]interface{}, len(secrets))
	for i, v := range secrets {
		items[i] = v
	}
	columns := []Column{
		{Header: "ID", Extractor: func(item interface{}) string { return string(*item.(*azsecrets.SecretProperties).ID)}},
		{Header: "Name", Extractor: func(item interface{}) string { return item.(*azsecrets.SecretProperties).ID.Name() }},
		{Header: "Version", Extractor: func(item interface{}) string { return item.(*azsecrets.SecretProperties).ID.Version() }},
		{Header: "Name", Extractor: func(item interface{}) string { return item.(*azsecrets.SecretProperties).ID.Name() }},
		{Header: "Version", Extractor: func(item interface{}) string { return item.(*azsecrets.SecretProperties).ID.Version() }},
		{Header: "Enabled", Extractor: func(item interface{}) string {
			return fmt.Sprintf("%t", *item.(*azsecrets.SecretProperties).Attributes.Enabled)
		},
		},
	}

	return FormatTable(items, columns)
}

