package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/mjsully/terraform-provider-upsnap/upsnap"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: upsnap.Provider,
	})
}
