// Entry point of the Nobl9 Terraform Provider.
package main

import (
	"context"
	"flag"
	"log"

	// "github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/nobl9/terraform-provider-nobl9/nobl9"
)

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs -provider-name=terraform-provider-nobl9

func main() {
	ctx := context.Background()
	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: nobl9.Provider,
	}

	if debugMode {
		opts.Debug = true
		opts.ProviderAddr = "nobl9.com/nobl9/nobl9"
	}

	providers := []func() tfprotov5.ProviderServer{
		// providerserver.NewProtocol5(nil), // Example terraform-plugin-framework provider
		opts.GRPCProviderFunc,
	}

	muxServer, err := tf5muxserver.NewMuxServer(ctx, providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf5server.ServeOpt
	if debugMode {
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
	}

	if err = tf5server.Serve(
		"registry.terraform.io/<namespace>/<provider_name>",
		muxServer.ProviderServer,
		serveOpts...,
	); err != nil {
		log.Fatal(err)
	}
}
