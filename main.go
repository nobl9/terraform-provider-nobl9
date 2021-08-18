package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/nobl9/terraform-provider-nobl9/nobl9"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: nobl9.Provider,
	})
}
