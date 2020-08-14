package main

import (
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v3/auth/noop"
	"github.com/micro/go-micro/v3/broker"
	"github.com/micro/go-micro/v3/broker/nats"
	"github.com/micro/go-micro/v3/client"
	"github.com/micro/go-micro/v3/config"
	"github.com/micro/go-micro/v3/registry"
	"github.com/micro/go-micro/v3/registry/etcd"
	"github.com/micro/go-micro/v3/router"
	regRouter "github.com/micro/go-micro/v3/router/registry"
	"github.com/micro/go-micro/v3/server"
	"github.com/micro/go-micro/v3/store"
	"github.com/micro/go-micro/v3/store/cockroach"

	"github.com/micro/micro/v3/profile"
	microAuth "github.com/micro/micro/v3/service/auth"
	microBroker "github.com/micro/micro/v3/service/broker"
	microClient "github.com/micro/micro/v3/service/client"
	microConfig "github.com/micro/micro/v3/service/config"
	microRegistry "github.com/micro/micro/v3/service/registry"
	microRouter "github.com/micro/micro/v3/service/router"
	microServer "github.com/micro/micro/v3/service/server"
	microStore "github.com/micro/micro/v3/service/store"
)

var prodServer = &profile.Profile{
	Name: "prod-server",
	Setup: func(ctx *cli.Context) error {
		microAuth.DefaultAuth = noop.NewAuth()
		microConfig.DefaultConfig, _ = config.NewConfig()
		setBroker(nats.NewBroker(broker.Addrs("nats-cluster")))
		setRegistry(etcd.NewRegistry(registry.Addrs("etcd-cluster")))
		microStore.DefaultStore = cockroach.NewStore(store.Nodes("postgresql://root@cockroachdb-cluster-public:26257?sslmode=disable"))
		return nil
	},
}

var prodClient = &profile.Profile{
	Name: "prod-client",
	Setup: func(ctx *cli.Context) error {
		microAuth.DefaultAuth = noop.NewAuth()
		microConfig.DefaultConfig, _ = config.NewConfig()
		microRegistry.DefaultRegistry.Init(registry.Addrs("micro-registry:8000"))
		return nil
	},
}

func init() {
	profile.Register("prod-server", prodServer)
	profile.Register("prod-client", prodClient)
}

func setRegistry(reg registry.Registry) {
	microRegistry.DefaultRegistry = reg
	microRouter.DefaultRouter = regRouter.NewRouter(router.Registry(reg))
	microServer.DefaultServer.Init(server.Registry(reg))
	microClient.DefaultClient.Init(client.Registry(reg))
}

func setBroker(b broker.Broker) {
	microBroker.DefaultBroker = b
	microClient.DefaultClient.Init(client.Broker(b))
	microServer.DefaultServer.Init(server.Broker(b))
}
