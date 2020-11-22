package profile

import (
	"github.com/micro/cli/v2"
	"github.com/micro/go-micro/v3/auth/noop"
	"github.com/micro/go-micro/v3/broker"
	"github.com/micro/go-micro/v3/broker/nats"
	"github.com/micro/go-micro/v3/registry/memory"
	"github.com/micro/go-micro/v3/store"
	"github.com/micro/go-micro/v3/store/cockroach"

	microAuth "github.com/micro/micro/v3/service/auth"
	microRouter "github.com/micro/micro/v3/service/router"
	"github.com/micro/micro/v3/service/router/kubernetes"
	microStore "github.com/micro/micro/v3/service/store"
)

var production = &Profile{
	Name: "production",
	Setup: func(ctx *cli.Context) error {
		microAuth.DefaultAuth = noop.NewAuth()
		setBroker(nats.NewBroker(broker.Addrs("nats-cluster")))
		setRegistry(memory.NewRegistry())
		microRouter.DefaultRouter = kubernetes.NewRouter()
		microStore.DefaultStore = cockroach.NewStore(store.Nodes("postgresql://root@cockroachdb-cluster-public:26257?sslmode=disable"))
		return nil
	},
}

func init() {
	Register("production", production)
}
