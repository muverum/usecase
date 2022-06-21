# Wrapping [swaggest/rest](https://github.com/swaggest/rest) -> Usecase



# UseCase

I've found that in the world of REST API building, the [swaggest/rest](https://github.com/swaggest/rest) library solves most of my problems:

1. Conform to the go-chi router specifications
2. Built around clean architecture concepts
3. Self documenting to a swagger document

The `UseCase` library wraps this into a Generic Struct functionality that allows your interactions to be defined 
more centrally within a package and be self contained:

```go
package usecase

import (
	"context"
	"encoding/json"
	"github.com/muverum/usecase"
	usecase2 "github.com/swaggest/usecase"
)

type ConcatenateRequest struct {
	_     struct{} `title:"concatenate this request with some data"`
	Input string   `json:"input" required:"true" minLength:"1"`
}

type ConcatenateResponse struct {
	_      struct{} `title:"response of concatenated values"`
	Output string   `json:"output"`
}

func (c *ConcatenateResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.Output)
}

func catUseCase() usecase.UseCaseFunc[ConcatenateRequest, *ConcatenateResponse] {
	return func(ctx context.Context, i ConcatenateRequest, o *ConcatenateResponse) error {
		o.Output = i.Input + "some-more-text"
		return nil
	}
}

func MakeCatUsecase() (usecase.UseCase[ConcatenateRequest, *ConcatenateResponse], error) {
	var decorationFunc = func(i *usecase2.IOInteractor) {
		i.SetTags("cat")
		i.SetTitle("Concatenate your request")
		i.SetDescription("Concatenates your request data with a fixed string")
	}

	return usecase.New(ConcatenateRequest{}, &ConcatenateResponse{}, catUseCase(), decorationFunc)
}

```

These can then be returned as either an `Interactor()` for use with the swaggest/rest library or as a 
`Handler()` for anything that takes an `http.Handler` like go-chi's `Method`

# SubRouting
[Working with the developer](https://github.com/swaggest/rest/issues/84) it was apparent that `Mount` was a can of 
worms unto itself. The author did point out that the `Route` methods did work

```go
// Set a subrouter that will use its own middleware (specifically the annotation for the parent collector
apiService.Route("/dog", func(r chi.Router) {
    r.Use(nethttp.AnnotateOpenAPI(apiService.OpenAPICollector, func(op *openapi3.Operation) error {
        op.Tags = []string{"Dogs"}

        return nil
    }))

    r.Method(http.MethodGet, "/walk/{place}/{times}", usecase2.MakeDogWalkUseCase(log.New(os.Stdout, "DOG-", 0)).Handler())
})
```

## Handler Method
In order to conform to the go-chi `Method` signature, we had to provide a `http.Handler`. This was relatively easy
since the `nethttp` library provided a new handler function for the interactor, so we just build the interactor
and then provide the handler. 

This allows for the `Usecase` to be the core business logic of any interaction. The router you're attaching to
just dictates whether you need a handler or an interactor.

## Usecase Middlewares

I'm going to leave the use case middlewares in place as it will allow the user to break up interactions
that may use the context into smaller, easier to test pieces. 