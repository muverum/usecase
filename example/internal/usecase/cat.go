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
