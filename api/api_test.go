package api

import (
	"encoding/json"
	"github.com/metrumresearchgroup/wrapt"
	"github.com/muverum/usecase"
	"github.com/muverum/usecase/example/nodes/dog"
	usecase2 "github.com/muverum/usecase/example/usecase"
	"github.com/muverum/usecase/node"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func testServer() *httptest.Server {
	api := New(3001, 3000)

	logger := log.New(os.Stdout, "EXAMPLE-", 0)

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
	dognode, err := dog.New(api.Server, logger)
	if err != nil {
		log.Fatal(err.Error())
	}

	//logger.Print(dognode.Routes())

	api.Nodes = []*node.Node{
		dognode,
	}

	_ = api.MountRoutes()
	return httptest.NewServer(api.Server)
}

func TestAPI_Listen(tt *testing.T) {

	server := testServer()
	defer server.Close()

	tests := []struct {
		Name          string
		Path          string
		Method        string
		Body          io.Reader
		wantErr       bool
		assertionFunc func(t *wrapt.T, res *http.Response)
	}{
		{
			Name:    "cat as expected",
			Path:    "/cat",
			Method:  "POST",
			Body:    strings.NewReader(`{ "input" : "banana"}`),
			wantErr: false,
			assertionFunc: func(t *wrapt.T, res *http.Response) {
				t.A.Equal(res.StatusCode, 200)
				var ConcatenateResponse usecase2.ConcatenateResponse
				//Unmarshalling only returns the string in this request
				t.A.Nil(json.NewDecoder(res.Body).Decode(&ConcatenateResponse.Output))
				t.A.Equal(ConcatenateResponse.Output, "bananasome-more-text")
			},
		},
		{
			Name:    "unmounted targets 404",
			Path:    "/wtf",
			Method:  "POST",
			Body:    strings.NewReader(`{ "input" : "banana"}`),
			wantErr: false,
			assertionFunc: func(t *wrapt.T, res *http.Response) {
				t.A.Equal(res.StatusCode, 404)
			},
		},
		{
			Name:    "mounted node doesn't respond at root",
			Path:    "/dog",
			Method:  "GET",
			Body:    nil,
			wantErr: false,
			assertionFunc: func(t *wrapt.T, res *http.Response) {
				t.A.Equal(res.StatusCode, 404)
			},
		},
		{
			Name:    "First Tree action responds appropriately",
			Path:    "/dog/walk/atlanta/4",
			Method:  "GET",
			Body:    nil,
			wantErr: false,
			assertionFunc: func(t *wrapt.T, res *http.Response) {
				t.A.Equal(res.StatusCode, 200)
				var output usecase2.DogWalkResponse
				t.A.Nil(json.NewDecoder(res.Body).Decode(&output))
				t.A.Equal(output.Times, 4)
				t.A.Equal(output.Walked, true)
			},
		},
		{
			Name:    "Second tree item responds (feed)",
			Path:    "/dog/feed",
			Method:  "POST",
			Body:    strings.NewReader(`{ "bowls" : 2 }`),
			wantErr: false,
			assertionFunc: func(t *wrapt.T, res *http.Response) {
				t.A.Equal(res.StatusCode, 200)
				var output usecase2.DogFeedResponse
				t.A.Nil(json.NewDecoder(res.Body).Decode(&output))
				t.A.Equal(output.Happy, true)
			},
		},
	}
	for _, test := range tests {
		tt.Run(test.Name, func(tt *testing.T) {
			t := wrapt.WrapT(tt)

			client := http.Client{}

			req, err := http.NewRequest(test.Method, server.URL+test.Path, test.Body)

			t.A.Equal(err != nil, test.wantErr)

			res, err := client.Do(req)

			t.A.Equal(err != nil, test.wantErr)

			if test.assertionFunc != nil {
				test.assertionFunc(t, res)
			}

		})
	}
}
