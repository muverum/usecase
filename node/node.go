package node

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/rest/nethttp"
	"github.com/swaggest/rest/web"
	"net/http"
	"strings"
)

type Route string

type Handler interface {
	Handler() http.Handler
}

type Node struct {
	//Root is the mountpoint for this node
	Root           string
	Tags           []string
	service        *web.Service
	Middleware     []func(next http.Handler) http.Handler
	DefaultOptions Handler
	// Tree reads as routePath -> map of http verb to its usecase
	Tree map[Route]map[string]Handler
}

func New(server *web.Service, options ...func(n *Node)) *Node {
	n := &Node{
		service: server,
	}

	for _, v := range options {
		v(n)
	}

	return n
}

func (a *Node) Use(middleware ...func(next http.Handler) http.Handler) {
	a.Middleware = append(a.Middleware, middleware...)
}

func (a *Node) Routes() string {
	sb := strings.Builder{}
	sb.WriteString(a.Root)
	sb.WriteString("\n")
	for route, v := range a.Tree {
		for verb, _ := range v {
			sb.WriteString(fmt.Sprintf("\t%s\t%s", verb, route))
		}
	}

	return sb.String()
}

func (a *Node) Validate() error {

	for route, _ := range a.Tree {
		//Error if not prefixed by /
		if !strings.HasPrefix(string(route), "/") {
			return fmt.Errorf("")
		}
	}

	return nil
}

func (a *Node) Mount() error {
	var err error
	if err = a.Validate(); err != nil {
		return err
	}
	a.service.Route(a.Root, func(r chi.Router) {
		//Define the middleware for this node if present
		if len(a.Middleware) > 0 {
			r.Use(a.Middleware...)
		}

		// Make sure the collector is wrapped accordingly
		if len(a.Tags) > 0 {
			r.Use(nethttp.AnnotateOpenAPI(a.service.OpenAPICollector, func(op *openapi3.Operation) error {
				op.Tags = a.Tags
				return nil
			}))
		}

		for route, v := range a.Tree {
			for verb, action := range v {
				switch verb {
				case http.MethodOptions:
					//apply the explicit options
					r.Method(verb, string(route), action.Handler())
				default:
					// apply Default if present
					if a.DefaultOptions != nil {
						r.Method(verb, string(route), a.DefaultOptions.Handler())
					}
					r.Method(verb, string(route), action.Handler())
				}
			}
		}
	})

	return nil
}
