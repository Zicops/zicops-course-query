package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	wasmhttp "github.com/nlepage/go-wasm-http-server"
	"github.com/zicops/zicops-course-query/global"
	graceful "gopkg.in/tylerb/graceful.v1" // see: https://github.com/tylerb/graceful
)

type maxPayloadHandler struct {
	handler http.Handler
	size    int64
}

// ServeHTTP uses MaxByteReader to limit the size of the input
func (handler *maxPayloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, handler.size)
	wasmhttp.Serve(handler.handler)
}

// CCBackendController ....
func CCBackendController(ctx context.Context, port int, errorChannel chan error) {
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

func serverHTTPRoutes(ctx context.Context, httpAddress string, handler http.Handler, errorChannel <-chan error) {
	defer global.WaitGroupServer.Done()
	// init graceful server
	serverGrace := &graceful.Server{
		Timeout: 10 * time.Second,
		//BeforeShutdown:    beforeShutDown,
		ShutdownInitiated: shutDownBackend,
		Server: &http.Server{
			Addr:    httpAddress,
			Handler: handler,
		},
	}
	stopChannel := serverGrace.StopChan()
	err := serverGrace.ListenAndServe()
	if err != nil {
		fmt.Println("CCController: Failed to start server : %s", err.Error())
	}
	fmt.Println("Backend is serving the routes.")
	for {
		// wait for the server to stop or be canceled
		select {
		case <-stopChannel:
			fmt.Println("CCController: Server shutdown at %s", time.Now())
			return
		case <-ctx.Done():
			fmt.Println("CCController: context done is called %s", time.Now())
			serverGrace.Stop(time.Second * 2)
		}
	}
}

func shutDownBackend() {
	fmt.Println("CCController: Shutting down server at %s", time.Now())
}
