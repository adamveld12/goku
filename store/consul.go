package store

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

var client *api.Client

func init() {
	var err error
	client, err = api.NewClient(api.DefaultConfig())

	if err != nil {
		panic(fmt.Sprintf("failed to initialize consul API\n%s\n", err.Error()))
	}
}

func RegisterWithServiceCatalog() error {
	return nil
}

func DeregisterWithServiceCatalog() error {
	return nil
}
