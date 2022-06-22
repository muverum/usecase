package main

import (
	"github.com/muverum/usecase"
	api2 "github.com/muverum/usecase/api"
	"github.com/muverum/usecase/example/internal/nodes/dog"
	usecase2 "github.com/muverum/usecase/example/internal/usecase"
	"github.com/muverum/usecase/node"
	"log"
	"net/http"
	"os"
)

func main() {

	api := api2.New(3001, 3000)

	//Build use case for top-level mount
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

	//Build a new node
	dognode, err := dog.New(api.Server, log.New(os.Stdout, "DOG-", 0))
	if err != nil {
		log.Fatal(err.Error())
	}

	api.Nodes = []*node.Node{
		dognode,
	}

	_ = api.Listen()
}
