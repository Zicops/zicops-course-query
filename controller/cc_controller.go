package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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
	handler.handler.ServeHTTP(w, r)
}

// CCBackendController ....
func CCBackendController(ctx context.Context, port int, errorChannel chan error, r *gin.Engine) {
	log.Infof("Initializing router and endpoints.")
	ccRouter, err := CCRouter(r)
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
		log.Fatalf("CCController: Failed to start server : %s", err.Error())
	}
	log.Infof("Backend is serving the routes.")
	for {
		// wait for the server to stop or be canceled
		select {
		case <-stopChannel:
			log.Infof("CCController: Server shutdown at %s", time.Now())
			return
		case <-ctx.Done():
			log.Infof("CCController: context done is called %s", time.Now())
			serverGrace.Stop(time.Second * 2)
		}
	}
}

func shutDownBackend() {
	log.Infof("CCController: Shutting down server at %s", time.Now())
}
