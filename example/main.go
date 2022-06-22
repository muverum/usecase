package main

import (
	"github.com/muverum/usecase"
	api2 "github.com/muverum/usecase/api"
	usecase2 "github.com/muverum/usecase/example/internal/usecase"
	"github.com/muverum/usecase/node"
	"log"
	"net/http"
	"os"
)

func main() {

	api := api2.New(3001, 3000)

	type thisRequest struct {
		Input string `json:"intpu"`
	}

	type thisResponse struct {
		Output string `json:"output"`
	}

	//We first build a use case and from it extract an interactor to hand to the router
	var catUseCase usecase.UseCase[usecase2.ConcatenateRequest, *usecase2.ConcatenateResponse]
	var err error
	if catUseCase, err = usecase2.MakeCatUsecase(); err != nil {
		log.Fatal(err.Error())
	}

	api.Actions = map[string]map[string]node.Handler{
		"/cat": {
			http.MethodPost: catUseCase,
		},
	}

	//What if we need a new subset of functionality that needs additional commonalities (besides) the middlewares
	//listed here, but don't make sense to add everywhere?

	dogNode := node.New(api.Server, func(n *node.Node) {

		var dogUseCase usecase.UseCase[usecase2.DogWalkRequest, *usecase2.DogWalkResponse]
		if dogUseCase, err = usecase2.MakeDogWalkUseCase(log.New(os.Stdout, "DOG-", 0)); err != nil {
			log.Fatal(err.Error())
		}

		n.Root = "/dog"
		n.Tags = []string{
			"dog",
		}

		n.Tree = map[node.Route]map[string]node.Handler{
			"/walk/{place}/{times}": {
				http.MethodGet: dogUseCase,
			},
		}
	})

	api.Nodes = []*node.Node{
		dogNode,
	}

	_ = api.Listen()
}
