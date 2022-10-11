package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	wasmhttp "github.com/nlepage/go-wasm-http-server"
	"github.com/zicops/zicops-course-query/global"
	// see: https://github.com/tylerb/graceful
)

type maxPayloadHandler struct {
	handler http.Handler
	size    int64
}

// CCBackendController ....
func CCBackendController(ctx context.Context, port int, errorChannel chan<- error) {
	fmt.Println("Initializing router and endpoints.")
	ccRouter, err := CCRouter()
	if err != nil {
		errorChannel <- err
		return
	}
	httpAddress := fmt.Sprintf(":%d", port)
	global.WaitGroupServer.Add(1)
	go serverHTTPRoutes(ctx, httpAddress, ccRouter, errorChannel)
}

func serverHTTPRoutes(ctx context.Context, httpAddress string, handler http.Handler, errorChannel chan<- error) {
	defer global.WaitGroupServer.Done()
	http.HandleFunc("/hello", func(res http.ResponseWriter, req *http.Request) {
		params := make(map[string]string)
		if err := json.NewDecoder(req.Body).Decode(&params); err != nil {
			panic(err)
		}

		res.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(res).Encode(map[string]string{
			"message": fmt.Sprintf("Hello %s!", params["name"]),
		}); err != nil {
			panic(err)
		}
	})

	wasmhttp.Serve(nil)

}

func shutDownBackend() {
	fmt.Println("CCController: Shutting down server at %s", time.Now())
}
