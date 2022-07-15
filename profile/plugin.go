package profile

import (
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v3/auth/noop"
	"github.com/micro/go-micro/v3/broker"
	"github.com/micro/go-micro/v3/broker/nats"
	"github.com/micro/go-micro/v3/client"
	"github.com/micro/go-micro/v3/config"

	microAuth "github.com/micro/micro/v3/service/auth"
	microClient "github.com/micro/micro/v3/service/client"
	microConfig "github.com/micro/micro/v3/service/config"
	k8sReg "github.com/micro/micro/v3/service/registry/kubernetes"
	"github.com/micro/micro/v3/service/router/static"
)

var production = &Profile{
	Name: "production",
	Setup: func(ctx *cli.Context) error {
		microAuth.DefaultAuth = noop.NewAuth()
		setBroker(nats.NewBroker(broker.Addrs("nats-cluster")))
		setRegistry(k8sReg.NewRegistry())
		microConfig.DefaultConfig, _ = config.NewConfig()
		microClient.DefaultClient.Init(client.Router(static.NewRouter()))
		return nil
	},
}

func init() {
	Register("production", production)
}
