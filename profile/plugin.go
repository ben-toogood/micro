package profile

import (
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v3/auth/noop"
	"github.com/micro/go-micro/v3/broker"
	"github.com/micro/go-micro/v3/broker/nats"
	"github.com/micro/go-micro/v3/client"
	"github.com/micro/go-micro/v3/store"
	"github.com/micro/go-micro/v3/store/cockroach"

	microAuth "github.com/micro/micro/v3/service/auth"
	microClient "github.com/micro/micro/v3/service/client"
	k8sReg "github.com/micro/micro/v3/service/registry/kubernetes"
	"github.com/micro/micro/v3/service/router/kubernetes"
	microStore "github.com/micro/micro/v3/service/store"
)

var production = &Profile{
	Name: "production",
	Setup: func(ctx *cli.Context) error {
		microAuth.DefaultAuth = noop.NewAuth()
		setBroker(nats.NewBroker(broker.Addrs("nats-cluster")))
		setRegistry(k8sReg.NewRegistry())
		microClient.DefaultClient.Init(client.Router(kubernetes.NewRouter()))
		microStore.DefaultStore = cockroach.NewStore(store.Nodes("postgresql://root@cockroachdb-cluster-public:26257?sslmode=disable"))
		return nil
	},
}

func init() {
	Register("production", production)
}
