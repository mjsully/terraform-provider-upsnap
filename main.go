package main

import (
	// "github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/mjsully/terraform-provider-upsnap/internal/provider"
)

var (
	version string = "dev"
)

// func main() {
// 	plugin.Serve(&plugin.ServeOpts{
// 		ProviderFunc: upsnap.Provider,
// 	})
// }

func main() {

	opts := providerserver.ServeOpts{
		// TODO: Update this string with the published name of your provider.
		// Also update the tfplugindocs generate command to either remove the
		// -provider-name flag or set its value to the updated provider name.
		Address: "registry.terraform.io/mjsully/upsnap",
	}

	err := providerserver.Serve(context.Background(), provider.New(version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}

}
