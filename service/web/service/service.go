// Package service provides web based micro services
package service

import (
	"context"
	"net/http"

	"github.com/micro/micro/v3/service"
)

// Service is a web service with service discovery built in
type Service interface {
	Handle(pattern string, handler http.Handler)
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	Run() error
}

// New returns a new web service
func New() Service {
	s := &webService{
		mux:     http.NewServeMux(),
		context: context.TODO(),
		Service: service.New(),
	}

	s.srv = s.genSrv()
	return s
}
