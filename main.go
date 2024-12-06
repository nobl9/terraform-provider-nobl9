// Entry point of the Nobl9 Terraform Provider.
package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5/tf5server"
	"github.com/hashicorp/terraform-plugin-mux/tf5muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/nobl9/terraform-provider-nobl9/internal/frameworkprovider"
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

	muxServer, err := tf5muxserver.NewMuxServer(
		ctx,
		newSDKProvider(),
		newFrameworkProvider(),
	)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf5server.ServeOpt
	name := "registry.terraform.io/nobl9/nobl9"
	if debugMode {
		name = "nobl9.com/nobl9/nobl9"
		serveOpts = append(serveOpts, tf5server.WithManagedDebug())
	}

	if err = tf5server.Serve(
		name,
		muxServer.ProviderServer,
		serveOpts...,
	); err != nil {
		log.Fatal(err)
	}
}

func newSDKProvider() func() tfprotov5.ProviderServer {
	return func() tfprotov5.ProviderServer {
		return schema.NewGRPCProviderServer(nobl9.Provider())
	}
}

func newFrameworkProvider() func() tfprotov5.ProviderServer {
	provider := frameworkprovider.New("TODO")
	return providerserver.NewProtocol5(provider)
}
