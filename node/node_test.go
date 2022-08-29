package node

import (
	"context"
	"github.com/metrumresearchgroup/wrapt"
	"github.com/muverum/usecase"
	"github.com/swaggest/rest/web"
	"net/http"
	"testing"
)

func sampleHandler(next http.Handler) http.Handler {
	return next
}

func additionalHandler(next http.Handler) http.Handler {
	return sampleHandler(next)
}

func TestNode_Use(tt *testing.T) {
	type fields struct {
		Root           string
		Tags           []string
		service        *web.Service
		Middleware     []func(next http.Handler) http.Handler
		DefaultOptions Handler
		Tree           map[Route]map[string]Handler
	}
	type args struct {
		middleware []func(next http.Handler) http.Handler
	}
	tests := []struct {
		name          string
		assertionFunc func(t *wrapt.T, middleware []func(next http.Handler) http.Handler)
		fields        fields
		args          args
	}{
		{
			name: "expected count",
			fields: fields{
				Root:    "/banana",
				service: web.DefaultService(),
				Middleware: []func(next http.Handler) http.Handler{
					sampleHandler,
				},
				Tree: nil,
			},
			args: args{
				middleware: []func(next http.Handler) http.Handler{
					additionalHandler,
				},
			},
			assertionFunc: func(t *wrapt.T, middleware []func(next http.Handler) http.Handler) {
				t.A.Len(middleware, 2)
			},
		},
	}
	for _, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			t := wrapt.WrapT(tt)
			a := &Node{
				Root:           test.fields.Root,
				Tags:           test.fields.Tags,
				service:        test.fields.service,
				Middleware:     test.fields.Middleware,
				DefaultOptions: test.fields.DefaultOptions,
				Tree:           test.fields.Tree,
			}

			a.Use(test.args.middleware...)

			if test.assertionFunc != nil {
				test.assertionFunc(t, a.Middleware)
			}

		})
	}
}

func ptr[I any](input I) *I {
	return &input
}

func TestNode_Routes(tt *testing.T) {
	uc1, _ := usecase.New[string, *string]("", ptr(""), func(ctx context.Context, input string, output *string) error { return nil }, nil, nil)

	type fields struct {
		Root           string
		Tags           []string
		service        *web.Service
		Middleware     []func(next http.Handler) http.Handler
		DefaultOptions Handler
		Tree           map[Route]map[string]Handler
	}
	tests := []struct {
		name          string
		assertionFunc func(t *wrapt.T, out string)
		fields        fields
	}{
		{
			name: "small route tree",
			assertionFunc: func(t *wrapt.T, out string) {
				t.A.Contains(out, "/iamlegend\n")
				t.A.Contains(out, "\t/but/not/really\tGET")
				t.A.Contains(out, "\t/but/not/really\tPATCH")
				t.A.Contains(out, "\t/but/not/really\tDELETE")
				t.A.Contains(out, "/but/not/really\tPOST")
				t.A.Contains(out, "/but/really\tGET")
			},
			fields: fields{
				Root:    "/iamlegend",
				service: web.DefaultService(),
				Tree: map[Route]map[string]Handler{
					"/but/not/really": {
						http.MethodPost:   uc1,
						http.MethodGet:    uc1,
						http.MethodPatch:  uc1,
						http.MethodDelete: uc1,
					},
					"/but/really": {
						http.MethodGet: uc1,
					},
				},
			},
		},
	}
	for _, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			t := wrapt.WrapT(tt)
			a := &Node{
				Root:           test.fields.Root,
				Tags:           test.fields.Tags,
				service:        test.fields.service,
				Middleware:     test.fields.Middleware,
				DefaultOptions: test.fields.DefaultOptions,
				Tree:           test.fields.Tree,
			}

			got := a.Routes()

			if test.assertionFunc != nil {
				test.assertionFunc(t, got)
			}
		})
	}
}
