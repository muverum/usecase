package usecase

import (
	"context"
	"errors"
	"github.com/swaggest/usecase"
)

type UseCaseFunc[I any, O any] func(ctx context.Context, input I, output O) error

// Middleware operates in the execution chain for interactor() and produces output / decisions to the
// chain based on the input, modifies output and returns a context from each stage to be adjusted
type Middleware[I any, O any] func(ctx context.Context, input I, output O) (context.Context, error)

type UseCase[I any, O any] struct {
	input   I
	output  O
	usecase UseCaseFunc[I, O]
	// Middleware are to be wrapped during the interaction phase such that they are executed in order
	// before the actual use case func is called.
	middleware        []Middleware[I, O]
	apiDecorationFunc func(IOInteractor *usecase.IOInteractor)
}

func (i UseCase[I, O]) Use(middlewares ...Middleware[I, O]) {
	i.middleware = append(i.middleware, middlewares...)
}

// interactor is a thin layer wrapping the generic around the interface expected by the underlying library
func (i UseCase[I, O]) interactor() usecase.Interact {
	return func(ctx context.Context, input, output interface{}) error {
		var in I
		var out O
		var ok bool

		if in, ok = input.(I); !ok {
			return errors.New("input could not be processed as generic")
		}

		if out, ok = output.(O); !ok {
			return errors.New("output could not be processed as generic")
		}

		// Now we'll generate a _new function_ based off of the middlewares
		outContext := ctx
		var outFn = func(ctx context.Context, input I, output O) error {

			for _, v := range i.middleware {
				var err error
				if outContext, err = v(outContext, input, output); err != nil {
					return err
				}
			}
			
			return i.usecase(outContext, in, out)
		}

		return outFn(outContext, in, out)
	}
}

// Interactor is the method that should be called outside the package to construct the interactor correctly
func (i UseCase[I, O]) Interactor() usecase.Interactor {
	u := usecase.NewIOI(i.input, i.output, i.interactor())
	pu := &u
	if i.apiDecorationFunc != nil {
		i.apiDecorationFunc(pu)
	}
	return u
}

func New[I any, O any](input I, output O, interactor UseCaseFunc[I, O], decorationFunc func(IOInteractor *usecase.IOInteractor), m ...Middleware[I, O]) UseCase[I, O] {
	return UseCase[I, O]{
		input:             input,
		output:            output,
		usecase:           interactor,
		apiDecorationFunc: decorationFunc,
		middleware:        m,
	}
}
