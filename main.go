// Entry point of the Nobl9 Terraform Provider.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/nobl9/terraform-provider-nobl9/internal/frameworkprovider"
	"github.com/nobl9/terraform-provider-nobl9/internal/version"
	"github.com/nobl9/terraform-provider-nobl9/nobl9"
)

// Run the docs generation tool, check its repository for more information on how it works and how docs
// can be customized.
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs -provider-name=terraform-provider-nobl9

func main() {
	ctx := context.Background()
	var (
		debugMode   bool
		showVersion bool
	)
	flag.BoolVar(&debugMode, "debug", false, "run the provider with support for debuggers like delve")
	flag.BoolVar(&showVersion, "version", false, "display version of the Provider")
	flag.Parse()

	if showVersion {
		fmt.Println(version.GetUserAgent())
		return
	}

	muxServer, err := tf6muxserver.NewMuxServer(
		ctx,
		newSDKProvider(ctx),
		newFrameworkProvider(),
	)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt
	name := "registry.terraform.io/nobl9/nobl9"
	if debugMode {
		name = "nobl9.com/nobl9/nobl9"
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	if err = tf6server.Serve(
		name,
		muxServer.ProviderServer,
		serveOpts...,
	); err != nil {
		log.Fatal(err)
	}
}

func newSDKProvider(ctx context.Context) func() tfprotov6.ProviderServer {
	return func() tfprotov6.ProviderServer {
		srv, _ := tf5to6server.UpgradeServer(ctx, func() tfprotov5.ProviderServer {
			return schema.NewGRPCProviderServer(nobl9.Provider())
		})
		return srv
	}
}

func newFrameworkProvider() func() tfprotov6.ProviderServer {
	provider := frameworkprovider.New()
	return providerserver.NewProtocol6(provider)
}
