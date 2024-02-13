// Entry point of the Nobl9 Terraform Provider.
package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/nobl9/terraform-provider-nobl9/nobl9"
)

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs -provider-name=terraform-provider-nobl9

func main() {
	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: nobl9.Provider,
	}

	if debugMode {
		opts.Debug = true
		opts.ProviderAddr = "nobl9.com/nobl9/nobl9"
		plugin.Serve(opts)
		return
	}

	plugin.Serve(opts)
}
