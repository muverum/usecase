package usecase

import (
	"context"
	"github.com/muverum/usecase"
	usecase2 "github.com/swaggest/usecase"
	"log"
)

type DogWalkRequest struct {
	_     struct{} `title:"DogWalkRequest"`
	Place string   `json:"place" requird:"true" path:"place"`
	Times int      `json:"times" path:"times" description:"The number of times to walk said dog"`
}

//Not expecting anything so a 204 is ok
type DogWalkResponse struct {
	Walked bool `json:"walked"`
	Times  int  `json:"times"`
}

func dogWalkStopMiddleware(logger *log.Logger) usecase.Middleware[DogWalkRequest, *DogWalkResponse] {
	return func(ctx context.Context, i DogWalkRequest, o *DogWalkResponse) (context.Context, error) {
		logger.Print("Stopping the dog")

		//The context is returned so that it can be appended to by sequential operations
		return ctx, nil
	}
}

func dogWalkSniffingMiddleeware(logger *log.Logger) usecase.Middleware[DogWalkRequest, *DogWalkResponse] {
	return func(ctx context.Context, i DogWalkRequest, o *DogWalkResponse) (context.Context, error) {
		logger.Print("Sniffing")

		ctx = context.WithValue(ctx, "times", i.Times)

		//The context is returned so that it can be appended to by sequential operations
		return ctx, nil
	}
}

func dogWalkWhereAmIMiddleware(logger *log.Logger) usecase.Middleware[DogWalkRequest, *DogWalkResponse] {
	return func(ctx context.Context, i DogWalkRequest, o *DogWalkResponse) (context.Context, error) {
		logger.Print("It looks like I am in ", i.Place)
		return ctx, nil
	}
}

// In a testing scenario, this would be the business logic function you'd want to test
func dogWalkUseCase() usecase.UseCaseFunc[DogWalkRequest, *DogWalkResponse] {
	return func(ctx context.Context, i DogWalkRequest, o *DogWalkResponse) error {
		o.Walked = true
		o.Times = i.Times
		return nil
	}
}

func MakeDogWalkUseCase(logger *log.Logger) (usecase.UseCase[DogWalkRequest, *DogWalkResponse], error) {

	var decorationFunc = func(i *usecase2.IOInteractor) {
		i.SetTags("dog")
		i.SetTitle("WalkDog")
		i.SetDescription("Walk the dog")
	}

	middleware := []usecase.Middleware[DogWalkRequest, *DogWalkResponse]{
		//The middleware functions can be tested with small footprints individually
		dogWalkWhereAmIMiddleware(logger),
		dogWalkSniffingMiddleeware(logger),
		dogWalkStopMiddleware(logger),
	}

	/*
		This will produce logged output of:
		DOG-It looks like I am in newyork (Variable from input)
		DOG-Sniffing
		DOG-Stopping the dog

		Per request
	*/

	return usecase.New(DogWalkRequest{}, &DogWalkResponse{}, dogWalkUseCase(), decorationFunc, middleware...)
}

// Dog Feed
type DogFeedRequest struct {
	Bowls int `json:"bowls" required:"true"`
}

type DogFeedResponse struct {
	Happy bool `json:"happy"`
}

func dogFeed(ctx context.Context, input DogFeedRequest, output *DogFeedResponse) error {
	output.Happy = input.Bowls >= 2
	return nil
}

func MakeDogFeedUseCase() (usecase.UseCase[DogFeedRequest, *DogFeedResponse], error) {

	decorator := func(i *usecase2.IOInteractor) {
		i.SetTags("dog")
		i.SetTitle("FeedDog")
		i.SetDescription("Feeds the dog X times and sees if it's happy")
	}

	return usecase.New(DogFeedRequest{}, &DogFeedResponse{}, dogFeed, decorator)
}
