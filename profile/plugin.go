package profile

import (
	"github.com/micro/go-micro/v3/auth/noop"
	"github.com/micro/go-micro/v3/broker"
	"github.com/micro/go-micro/v3/registry"
	"github.com/micro/go-micro/v3/store"
	"github.com/micro/go-plugins/broker/nats/v3"
	"github.com/micro/go-plugins/registry/etcd/v3"
	"github.com/micro/go-plugins/store/cockroach/v3"
	"github.com/urfave/cli/v2"

	microAuth "github.com/micro/micro/v3/service/auth"
	microStore "github.com/micro/micro/v3/service/store"
)

var production = &Profile{
	Name: "production",
	Setup: func(ctx *cli.Context) error {
		microAuth.DefaultAuth = noop.NewAuth()
		SetupBroker(nats.NewBroker(broker.Addrs("nats-cluster")))
		SetupRegistry(etcd.NewRegistry(registry.Addrs("etcd-cluster")))
		microStore.DefaultStore = cockroach.NewStore(store.Nodes("postgresql://root@cockroachdb-cluster-public:26257?sslmode=disable"))
		return nil
	},
}

func init() {
	Register("production", production)
}
