package dog

import (
	"github.com/muverum/usecase"
	usecase2 "github.com/muverum/usecase/example/internal/usecase"
	"github.com/muverum/usecase/node"
	"github.com/swaggest/rest/web"
	"log"
	"net/http"
)

func New(service *web.Service, logger *log.Logger) (*node.Node, error) {
	var dogUseCase usecase.UseCase[usecase2.DogWalkRequest, *usecase2.DogWalkResponse]
	var err error
	if dogUseCase, err = usecase2.MakeDogWalkUseCase(logger); err != nil {
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
	}

	return n, nil
}
