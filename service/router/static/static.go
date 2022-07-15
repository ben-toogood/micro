// Package static is a static router which returns the service name as the address + port
package static

import (
	"fmt"
	"strings"

	"github.com/micro/go-micro/v3/router"
)

var (
	// DefaulPort is the port to append where nothing is set
	DefaultPort = 8080
)

// NewRouter returns an initialized static router
func NewRouter() router.Router {
	return &static{}
}

type static struct {
}

func (s *static) Init(opts ...router.Option) error {
	return nil
}

func (s *static) Options() router.Options {
	return router.Options{}
}

func (s *static) Table() router.Table {
	return nil
}

func (s *static) Lookup(service string, opts ...router.LookupOption) ([]router.Route, error) {
	return []router.Route{
		{
			Service: service,
			Address: fmt.Sprintf("%s:%d", strings.ReplaceAll(service, ".", "-"), DefaultPort),
			Gateway: "",
			Network: "",
			Router:  "",
		},
	}, nil
}

func (s *static) Watch(opts ...router.WatchOption) (router.Watcher, error) {
	return nil, nil
}

func (s *static) Close() error {
	return nil
}

func (s *static) String() string {
	return "static"
}
