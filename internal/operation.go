package internal

import (
	"fmt"
)

type Operation int

const (
	ListVersions Operation = iota
	GetPasswords
	ListVersionAndGetPasswords
	EditMetaData
	DeleteSecret
)

type OperationData struct {
	Name        string
	Description string
	Delimiter   string
}

var allOperations = map[Operation]OperationData{
	ListVersions:               {"list-versions", "List versions for selected items", ";"},
	GetPasswords:               {"get-passwords", "Get passwords for selected items", ";"},
	ListVersionAndGetPasswords: {"list-version-get-password", "List versions and get passwords for selected items", ";"},
	EditMetaData:               {"edit-meta", "Edit meta data for selected items in $EDITOR", ";"},
	DeleteSecret:               {"delete-secret", "Delete selected secret and all of its versions", ";"},
}

func (operation Operation) Data() OperationData {
	return allOperations[operation]
}

func StringToOperation(name string) (Operation, error) {
	for item, data := range allOperations {
		if data.Name == name {
			return item, nil
		}
	}
	return 0, fmt.Errorf("unknown operation: %q", name)
}
