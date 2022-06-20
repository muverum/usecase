# Wrapping [swaggest/rest](https://github.com/swaggest/rest) -> Usecase

I've found that in the world of REST API building, the [swaggest/rest](https://github.com/swaggest/rest) library solves most of my problems:

1. Conform to the go-chi router specifications
2. Built around clean architecture concepts
3. Self documenting to a swagger document

## Sharp Edges

I recently used the library to build an administrative API at work and ran into the fundamental problem. Let's say you
start with something simple like this:

```go
func main() {

	swagger := chi.NewMux()

	authenticated := web.DefaultService()

	authenticated.OpenAPI.Info.Title = "My magical API"
	authenticated.OpenAPI.Info.WithDescription("How we get things doneded.")
	authenticated.OpenAPI.Info.Version = "v0.0.1"
	//How to set this? O.o
	authenticated.OpenAPI.WithSecurity(map[string][]string{
		"security": []string{
			"bearerAuth",
		},
	})

	authenticated.OpenAPI.WithComponents(openapi3.Components{
		SecuritySchemes: &openapi3.ComponentsSecuritySchemes{
			map[string]openapi3.SecuritySchemeOrRef{
				"bearerAuth": openapi3.SecuritySchemeOrRef{
					SecurityScheme: &openapi3.SecurityScheme{
						HTTPSecurityScheme: &openapi3.HTTPSecurityScheme{
							Scheme:       "bearer",
							BearerFormat: safety.Ptr("JWT"),
						},
					},
				},
			},
		},
	})

	// Setup middlewares.
	authenticated.Wrap(
		gzip.Middleware, // Response compression with support for direct gzip pass through.
	)
	
	authenticated.Use(
		    all,
			the,
			middlewares,
		)
	
	authenticated.Get("/this", new usecase)

	

	go func() {
		logrus.Info("starting swagger server")
		http.ListenAndServe(":3000", swagger)
	}()

	logrus.Info("starting api server")
	http.ListenAndServe(":8001", authenticated)

}

// This is basically just used to make sure a _separate_ router is handling the swagger presentation
// since globally the `authenticated` router processes tokens.
func Docs(s chi.Router, pattern string, swgui func(title, schemaURL, basePath string) http.Handler, collector *openapi.Collector, spec *openapi3.Spec) {
	pattern = strings.TrimRight(pattern, "/")
	s.Method(http.MethodGet, pattern+"/openapi.json", collector)
	s.Mount(pattern, swgui(spec.Info.Title, pattern+"/openapi.json", pattern))
}
```

In this scenario, we apply a bunch of middlewares at the top (JWT processing, user load etc) into the context.
What if, however, you now wanted to mount a _sub_ router?

## Issues with Collection
If you create a sub-router _sharing_ the collector, you'll run into a few potential problems

### Invalid Docs
```go
authenticated := web.DefaultService()
sub1 := web.DefaultService(func(s *web.Service, initialized bool) {
    s.OpenAPICollector = authenticated.OpenAPICollector
})
authenticated.Mount("/there", sub1)
```

You'll have a *functionally correct* API with incorrect documentation. The swagger doc will only see `/` as the
root of whatever is mounted ot `sub1`. 

### Invalid Mountpoints
You _could_ then say

> In order to make the docs right, every sub router and its routes will need a full path definition

Whhich solves the documentation problem, but then introduces a functional problem. Now the documentation in openapi
will accurately reflect your _intent_ but not the implementation. Now functionally you'll have hugely long and
repetitive routes, and when you try to access the routes from openapi's UI, you'll just get `404`

## The Solution
What I did to work around this was to operate on the assumption `Middleware` has multiple meanings. 

1. Globally, they can be applied to the Chi wrapped constructs to perform _global, http level_ operations
2. `Middleware` could also be scoped to a `UseCase` which is a new type added in this library

This allows you to:

1. Maintain global API level middlewares (if desired)
2. Define middlewares as return funcs for various common features
3. Apply them by usecase where necessary

In this model you can still live within the scope of a global router, but have the flexibility of defining common
interactions on a resource-by-resource basis.

### What are you talking about
Let's say that you have a resource that touches a cloud provider (AWS for exmaple) that _must_ use STS 
to perform its work:

* STS access is short lived. If you build the client at startup, its access will expire over time

You _could_ write a goroutine-based process to constantly update that client _but if you're not getting requests to
process_ that doesn't make sense. In this model you would have middleware to:

1. Build the client from a config
2. Add it to the context

The interactor that gets generated can then read from the context and use the retrieved client to
perform whatever is necesary to produce the output