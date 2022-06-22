package api

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/muverum/usecase/node"
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/rest/openapi"
	"github.com/swaggest/rest/response/gzip"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v4emb"
	"net/http"
	"strings"
)

type API struct {
	Server *web.Service
	Nodes  []*node.Node
	//Global middleware
	Middleware []func(next http.Handler) http.Handler
	Wraps      []func(next http.Handler) http.Handler
	Actions    map[string]map[string]node.Handler
	// Port Defines the listening TCP Port for this when started
	Ports struct {
		API     int
		Swagger int
	}
}

func Docs(s chi.Router, pattern string, swgui func(title, schemaURL, basePath string) http.Handler, collector *openapi.Collector, spec *openapi3.Spec) {
	pattern = strings.TrimRight(pattern, "/")
	s.Method(http.MethodGet, pattern+"/openapi.json", collector)
	s.Mount(pattern, swgui(spec.Info.Title, pattern+"/openapi.json", pattern))
}

func New(apiPort int, swaggerPort int, options ...func(s *web.Service, initialized bool)) *API {

	server := web.DefaultService(options...)

	return &API{
		Server: server,
		Middleware: []func(next http.Handler) http.Handler{
			middleware.RequestID,
			middleware.Logger,
			middleware.Recoverer,
		},
		Wraps: []func(next http.Handler) http.Handler{
			gzip.Middleware,
		},
		Ports: struct {
			API     int
			Swagger int
		}{API: apiPort, Swagger: swaggerPort},
	}
}

func (a *API) Listen() error {

	a.Server.Wrap(a.Wraps...)
	a.Server.Use(a.Middleware...)

	//Mount top level actions
	for route, v := range a.Actions {
		for method, h := range v {
			a.Server.Method(method, route, h.Handler())
		}
	}
	
	r := chi.NewRouter()
	Docs(r, "/swagger", swgui.New, a.Server.OpenAPICollector, a.Server.OpenAPI)

	var err error
	for _, v := range a.Nodes {
		if err = v.Mount(); err != nil {
			return err
		}
	}
	go func() {
		_ = http.ListenAndServe(fmt.Sprintf(":%d", a.Ports.Swagger), r)
	}()

	return http.ListenAndServe(fmt.Sprintf(":%d", a.Ports.API), a.Server)
}
