package dog

import (
	"github.com/muverum/usecase"
	usecase2 "github.com/muverum/usecase/example/usecase"
	log2 "github.com/muverum/usecase/log"
	"github.com/muverum/usecase/node"
	"github.com/swaggest/rest/web"
	"log"
	"net/http"
)

func New(service *web.Service, logger *log.Logger) (*node.Node, error) {
	var dogUseCase usecase.UseCase[usecase2.DogWalkRequest, *usecase2.DogWalkResponse]
	var err error
	if dogUseCase, err = usecase2.MakeDogWalkUseCase(logger, log2.NewLogWrapper(logger)); err != nil {
		return nil, err
	}

	var dogFeedUseCase usecase.UseCase[usecase2.DogFeedRequest, *usecase2.DogFeedResponse]
	if dogFeedUseCase, err = usecase2.MakeDogFeedUseCase(log2.NewLogWrapper(logger)); err != nil {
		return nil, err
	}

	n := node.New(service)
	n.Root = "/dog"
	n.Tags = []string{
		"dog",
		"canine",
	}
	n.Tree = map[node.Route]map[string]node.Handler{
		"/walk/{place}/{times}": {
			http.MethodGet: dogUseCase,
		},
		"/feed": {
			http.MethodPost: dogFeedUseCase,
		},
	}

	return n, nil
}
