package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-micro/v3/logger"
	"github.com/micro/go-micro/v3/registry"
	"github.com/micro/go-micro/v3/server"
	maddr "github.com/micro/go-micro/v3/util/addr"
	"github.com/micro/go-micro/v3/util/backoff"
	mnet "github.com/micro/go-micro/v3/util/net"
	signalutil "github.com/micro/go-micro/v3/util/signal"

	"github.com/micro/micro/v3/service"
	muregistry "github.com/micro/micro/v3/service/registry"
	muserver "github.com/micro/micro/v3/service/server"
)

type webService struct {
	context context.Context
	mux     *http.ServeMux
	srv     *registry.Service

	running bool
	exit    chan chan error

	sync.RWMutex
	*service.Service
}

func (s *webService) genSrv() *registry.Service {
	var host string
	var port string
	var err error

	// default host:port
	if addr := muserver.DefaultServer.Options().Address; len(addr) > 0 {
		host, port, err = net.SplitHostPort(addr)
		if err != nil {
			logger.Fatal(err)
		}
	}

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if adv := muserver.DefaultServer.Options().Advertise; len(adv) > 0 {
		host, port, err = net.SplitHostPort(adv)
		if err != nil {
			logger.Fatal(err)
		}
	}

	addr, err := maddr.Extract(host)
	if err != nil {
		logger.Fatal(err)
	}

	if strings.Count(addr, ":") > 0 {
		addr = "[" + addr + "]"
	}

	return &registry.Service{
		Name:    muserver.DefaultServer.Options().Name,
		Version: muserver.DefaultServer.Options().Version,
		Nodes: []*registry.Node{{
			Id:      muserver.DefaultServer.Options().Id,
			Address: fmt.Sprintf("%s:%s", addr, port),
		}},
	}
}

func (s *webService) run(exit chan bool) {
	s.RLock()

	ttl := muserver.DefaultServer.Options().RegisterTTL
	if ttl <= time.Duration(0) {
		s.RUnlock()
		return
	}

	t := time.NewTicker(ttl)
	s.RUnlock()

	for {
		select {
		case <-t.C:
			s.register()
		case <-exit:
			t.Stop()
			return
		}
	}
}

func (s *webService) register() error {
	s.Lock()
	defer s.Unlock()

	if s.srv == nil {
		return nil
	}

	// service node need modify, node address maybe changed
	srv := s.genSrv()
	srv.Endpoints = s.srv.Endpoints
	s.srv = srv

	var regErr error

	// try three times if necessary
	for i := 0; i < 3; i++ {
		// attempt to register
		ttl := muserver.DefaultServer.Options().RegisterTTL
		if err := muregistry.Register(s.srv, registry.RegisterTTL(ttl)); err != nil {
			// set the error
			regErr = err
			// backoff then retry
			time.Sleep(backoff.Do(i + 1))
			continue
		}
		// success so nil error
		regErr = nil
		break
	}

	return regErr
}

func (s *webService) deregister() error {
	s.Lock()
	defer s.Unlock()

	if s.srv == nil {
		return nil
	}

	return muregistry.Deregister(s.srv)
}

func (s *webService) start() error {
	s.Lock()
	defer s.Unlock()

	if s.running {
		return nil
	}

	addr := muserver.DefaultServer.Options().Address
	l, err := mnet.Listen(addr, func(addr string) (net.Listener, error) {
		return net.Listen("tcp", addr)
	})
	if err != nil {
		return err
	}

	muserver.DefaultServer.Init(server.Address(l.Addr().String()))

	srv := s.genSrv()
	srv.Endpoints = s.srv.Endpoints
	s.srv = srv

	var h http.Handler = s.mux
	var r sync.Once

	// register the static html dir
	r.Do(func() {
		dir, _ := os.Getwd()
		static := filepath.Join(dir, "static")

		_, err := os.Stat(static)
		if err == nil {
			if logger.V(logger.InfoLevel, logger.DefaultLogger) {
				logger.Infof("Enabling static file serving from %s", static)
			}
			s.mux.Handle("/", http.FileServer(http.Dir(static)))
		}
	})

	httpSrv := &http.Server{
		Handler: h,
	}
	go httpSrv.Serve(l)

	s.exit = make(chan chan error, 1)
	s.running = true

	go func() {
		ch := <-s.exit
		ch <- l.Close()
	}()

	if logger.V(logger.InfoLevel, logger.DefaultLogger) {
		logger.Infof("Listening on %v", l.Addr().String())
	}
	return nil
}

func (s *webService) stop() error {
	s.Lock()
	defer s.Unlock()

	if !s.running {
		return nil
	}

	ch := make(chan error, 1)
	s.exit <- ch
	s.running = false

	if logger.V(logger.InfoLevel, logger.DefaultLogger) {
		logger.Info("Stopping")
	}

	return <-ch
}

func (s *webService) Handle(pattern string, handler http.Handler) {
	var seen bool
	s.RLock()
	for _, ep := range s.srv.Endpoints {
		if ep.Name == pattern {
			seen = true
			break
		}
	}
	s.RUnlock()

	// if its unseen then add an endpoint
	if !seen {
		s.Lock()
		s.srv.Endpoints = append(s.srv.Endpoints, &registry.Endpoint{
			Name: pattern,
		})
		s.Unlock()
	}

	// register the handler
	s.mux.Handle(pattern, handler)
}

func (s *webService) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {

	var seen bool
	s.RLock()
	for _, ep := range s.srv.Endpoints {
		if ep.Name == pattern {
			seen = true
			break
		}
	}
	s.RUnlock()

	if !seen {
		s.Lock()
		s.srv.Endpoints = append(s.srv.Endpoints, &registry.Endpoint{
			Name: pattern,
		})
		s.Unlock()
	}

	s.mux.HandleFunc(pattern, handler)
}

func (s *webService) Run() error {
	if err := s.start(); err != nil {
		return err
	}

	if err := s.register(); err != nil {
		return err
	}

	// start reg loop
	ex := make(chan bool)
	go s.run(ex)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, signalutil.Shutdown()...)

	select {
	// wait on kill signal
	case sig := <-ch:
		if logger.V(logger.InfoLevel, logger.DefaultLogger) {
			logger.Infof("Received signal %s", sig)
		}
	// wait on context cancel
	case <-s.context.Done():
		if logger.V(logger.InfoLevel, logger.DefaultLogger) {
			logger.Info("Received context shutdown")
		}
	}

	// exit reg loop
	close(ex)

	if err := s.deregister(); err != nil {
		return err
	}

	return s.stop()
}
