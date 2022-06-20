package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	usecase2 "github.com/muverum/usecase/example/internal/usecase"
	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/rest/openapi"
	"github.com/swaggest/rest/response/gzip"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v4emb"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {

	swagger := chi.NewMux()

	apiService := web.DefaultService()

	apiService.OpenAPI.Info.Title = "My magical API"
	apiService.OpenAPI.Info.WithDescription("How we get things doneded.")
	apiService.OpenAPI.Info.Version = "v0.0.1"
	//How to set this? O.o
	apiService.OpenAPI.WithSecurity(map[string][]string{
		"security": []string{
			"bearerAuth",
		},
	})

	jwt := "JWT"

	// If you want to specify that tokens are required for the docs:
	apiService.OpenAPI.WithComponents(openapi3.Components{
		SecuritySchemes: &openapi3.ComponentsSecuritySchemes{
			map[string]openapi3.SecuritySchemeOrRef{
				"bearerAuth": openapi3.SecuritySchemeOrRef{
					SecurityScheme: &openapi3.SecurityScheme{
						HTTPSecurityScheme: &openapi3.HTTPSecurityScheme{
							Scheme:       "bearer",
							BearerFormat: &jwt,
						},
					},
				},
			},
		},
	})

	// Setup middlewares.
	apiService.Wrap(
		gzip.Middleware, // Response compression with support for direct gzip pass through.
	)

	apiService.Use(
		middleware.RequestID,
		middleware.Logger,
		middleware.Recoverer,
	)

	type thisRequest struct {
		Input string `json:"intpu"`
	}

	type thisResponse struct {
		Output string `json:"output"`
	}

	//We first build a use case and from it extract an interactor to hand to the router
	apiService.Post("/cat", usecase2.MakeCatUsecase().Interactor())

	//What if we need a new subset of functionality that needs additional commonalities (besides) the middlewares
	//listed here, but don't make sense to add everywhere?

	// In this usecase constructor we actually set middleware
	apiService.Get("/dog/walk/{place}/{times}", usecase2.MakeDogWalkUseCase(log.New(os.Stdout, "DOG-", 0)).Interactor())

	Docs(swagger, "/swagger", swgui.New, apiService.OpenAPICollector, apiService.OpenAPI)

	go func() {
		http.ListenAndServe(":3000", swagger)
	}()

	http.ListenAndServe(":8001", apiService)

}

// This is basically just used to make sure a _separate_ router is handling the swagger presentation
// since globally the `authenticated` router processes tokens.
func Docs(s chi.Router, pattern string, swgui func(title, schemaURL, basePath string) http.Handler, collector *openapi.Collector, spec *openapi3.Spec) {
	pattern = strings.TrimRight(pattern, "/")
	s.Method(http.MethodGet, pattern+"/openapi.json", collector)
	s.Mount(pattern, swgui(spec.Info.Title, pattern+"/openapi.json", pattern))
}
