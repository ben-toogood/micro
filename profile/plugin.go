package profile

import (
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v3/auth/noop"
	"github.com/micro/go-micro/v3/broker"
	"github.com/micro/go-micro/v3/broker/nats"
	"github.com/micro/go-micro/v3/config"
	"github.com/micro/go-micro/v3/registry"
	"github.com/micro/go-micro/v3/registry/etcd"
	"github.com/micro/go-micro/v3/store"
	"github.com/micro/go-micro/v3/store/cockroach"

	microAuth "github.com/micro/micro/v3/service/auth"
	microConfig "github.com/micro/micro/v3/service/config"
	microRegistry "github.com/micro/micro/v3/service/registry"
	microStore "github.com/micro/micro/v3/service/store"
)

var production = &Profile{
	Name: "production",
	Setup: func(ctx *cli.Context) error {
		microAuth.DefaultAuth = noop.NewAuth()
		setBroker(nats.NewBroker(broker.Addrs("nats-cluster")))
		setRegistry(etcd.NewRegistry(registry.Addrs("etcd-cluster")))
		microStore.DefaultStore = cockroach.NewStore(store.Nodes("postgresql://root@cockroachdb-cluster-public:26257?sslmode=disable"))
		return nil
	},
}

var prodServer = &Profile{
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

var prodClient = &Profile{
	Name: "prod-client",
	Setup: func(ctx *cli.Context) error {
		microAuth.DefaultAuth = noop.NewAuth()
		setBroker(nats.NewBroker(broker.Addrs("nats-cluster")))
		microRegistry.DefaultRegistry.Init(registry.Addrs("micro-registry:8000"))
		return nil
	},
}

func init() {
	Register("production", production)
	Register("prod-server", prodServer)
	Register("prod-client", prodClient)
}
