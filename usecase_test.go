package usecase

import (
	"context"
	"errors"
	"github.com/metrumresearchgroup/wrapt"
	"github.com/swaggest/usecase"
	"strconv"
	"testing"
)

func TestUseCase_interactor(tt *testing.T) {

	type Input struct {
		Text   string `json:"string" minLength:"4"`
		Number int    `json:"number" required:"true"`
	}

	type Output struct {
		Message string `json:"message"`
		Counter *int   `json:"counter"`
	}

	incrementor := func(ctx context.Context, input Input, output *Output) (context.Context, error) {
		counter := ctx.Value("counter").(int)
		counter++
		return context.WithValue(ctx, "counter", counter), nil
	}

	type fields struct {
		input             Input
		output            *Output
		usecase           UseCaseFunc[Input, *Output]
		middleware        []Middleware[Input, *Output]
		apiDecorationFunc func(IOInteractor *usecase.IOInteractor)
	}
	tests := []struct {
		name          string
		fields        fields
		assertionFunc func(t *wrapt.T, err error, i Input, o *Output)
		ctxFunc       func() context.Context
		wantErr       bool
	}{
		{
			name:    "as expected",
			wantErr: false,
			fields: fields{
				input: Input{
					Text:   "oh hai there",
					Number: 5,
				},
				output: &Output{},
				usecase: func(ctx context.Context, input Input, output *Output) error {
					output.Message = input.Text + " " + strconv.Itoa(input.Number)
					return nil
				},
			},
			assertionFunc: func(t *wrapt.T, err error, i Input, o *Output) {
				t.A.Equal(o.Message, "oh hai there 5")
			},
			ctxFunc: func() context.Context {
				return context.Background()
			},
		},
		{
			name:    "error encountered in use case func",
			wantErr: true,
			fields: fields{
				input: Input{
					Text:   "oh hai there",
					Number: 5,
				},
				output: &Output{},
				usecase: func(ctx context.Context, input Input, output *Output) error {
					return errors.New("oh noes")
				},
			},
			assertionFunc: func(t *wrapt.T, err error, i Input, o *Output) {
				t.A.Equal(err.Error(), "oh noes")
				t.A.Empty(o)
			},
			ctxFunc: func() context.Context {
				return context.Background()
			},
		},
		{
			name:    "middlewares applying",
			wantErr: false,
			fields: fields{
				middleware: []Middleware[Input, *Output]{
					// Middleware run 4 times should increment 4 times
					incrementor,
					incrementor,
					incrementor,
					incrementor,
				},
				input: Input{
					Text:   "oh hai there",
					Number: 5,
				},
				output: &Output{},
				usecase: func(ctx context.Context, input Input, output *Output) error {
					output.Message = input.Text + " " + strconv.Itoa(input.Number)
					//Get the CTR out of the context
					counter := ctx.Value("counter").(int)
					output.Counter = &counter
					return nil
				},
			},

			assertionFunc: func(t *wrapt.T, err error, i Input, o *Output) {
				t.A.NotEmpty(o)
				t.A.NotNil(o.Counter)
				t.A.Equal(*o.Counter, 4)
			},
			ctxFunc: func() context.Context {
				return context.WithValue(context.Background(), "counter", 0)
			},
		},
		{
			name:    "non pointer provided as output should return an error",
			wantErr: true,
			fields: fields{
				input: Input{
					Text:   "oh hai there",
					Number: 5,
				},
				output: &Output{},
				usecase: func(ctx context.Context, input Input, output *Output) error {
					return errors.New("oh noes")
				},
			},
			assertionFunc: func(t *wrapt.T, err error, i Input, o *Output) {
				t.A.Equal(err.Error(), "oh noes")
				t.A.Empty(o)
			},
			ctxFunc: func() context.Context {
				return context.Background()
			},
		},
	}
	for _, test := range tests {
		tt.Run(test.name, func(tt *testing.T) {
			t := wrapt.WrapT(tt)
			t.R.NotNil(test.ctxFunc)
			i := UseCase[Input, *Output]{
				input:             test.fields.input,
				output:            test.fields.output,
				usecase:           test.fields.usecase,
				middleware:        test.fields.middleware,
				apiDecorationFunc: test.fields.apiDecorationFunc,
			}
			interactor := i.interactor()

			//Run the interactor
			err := interactor.Interact(test.ctxFunc(), test.fields.input, test.fields.output)

			t.A.Equal(err != nil, test.wantErr)

			if test.assertionFunc != nil {
				test.assertionFunc(t, err, test.fields.input, test.fields.output)
			}
		})
	}
}

func TestNewWithoutPointer(t *testing.T) {
	type Input struct {
		Input string `json:"string"`
	}

	type Output struct {
		Output string `json:"output"`
	}
	type args struct {
		input          Input
		output         Output
		interactor     UseCaseFunc[Input, Output]
		decorationFunc func(IOInteractor *usecase.IOInteractor)
		m              []Middleware[Input, Output]
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Should fail if a pointer value isn't provided",
			args: args{
				input:  Input{},
				output: Output{},
				interactor: func(ctx context.Context, input Input, output Output) error {
					return nil
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.input, tt.args.output, tt.args.interactor, tt.args.decorationFunc, nil, tt.args.m...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestNewWithPointer(t *testing.T) {
	type Input struct {
		Input string `json:"string"`
	}

	type Output struct {
		Output string `json:"output"`
	}
	type args struct {
		input          Input
		output         *Output
		interactor     UseCaseFunc[Input, *Output]
		decorationFunc func(IOInteractor *usecase.IOInteractor)
		m              []Middleware[Input, *Output]
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Should fail if a pointer value isn't provided",
			args: args{
				input:  Input{},
				output: &Output{},
				interactor: func(ctx context.Context, input Input, output *Output) error {
					return nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.args.input, tt.args.output, tt.args.interactor, tt.args.decorationFunc, nil, tt.args.m...)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
