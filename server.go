package main

import (
	"context"
	"fmt"
	"os/signal"
	"strconv"
	"syscall"

	"os"

	log "github.com/sirupsen/logrus"
	"github.com/zicops/zicops-course-query/config"
	"github.com/zicops/zicops-course-query/controller"
	"github.com/zicops/zicops-course-query/global"
	"github.com/zicops/zicops-course-query/lib/db/cassandra"
)

const defaultPort = "8080"

func main() {
	//os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "zicops-cc.json")
	log.Infof("Starting zicops course query service")
	ctx, cancel := context.WithCancel(context.Background())

	cassConfig := config.NewCassandraConfig()
	cassSession, err := cassandra.New(cassConfig)
	if err != nil {
		log.Errorf("Error connecting to cassandra: %s", err)
		log.Infof("zicops course creator intialization failed")
	}

	global.CTX = ctx
	global.CassSession = cassSession
	global.Cancel = cancel
	log.Infof("zicops course query intialization complete")
	portFromEnv := os.Getenv("PORT")
	port, err := strconv.Atoi(portFromEnv)

	if err != nil {
		port = 8091
	}
	bootUPErrors := make(chan error, 1)
	go monitorSystem(cancel, bootUPErrors)
	controller.CCBackendController(ctx, port, bootUPErrors)
	err = <-bootUPErrors
	if err != nil {
		log.Errorf("There is an issue starting backend server for course query: %v", err.Error())
		global.WaitGroupServer.Wait()
		os.Exit(1)
	}
	log.Infof("course query server started successfully.")
}

func monitorSystem(cancel context.CancelFunc, errorChannel chan error) {
	holdSignal := make(chan os.Signal, 1)
	signal.Notify(holdSignal, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	// if system throw any termination stuff let channel handle it and cancel
	<-holdSignal
	cancel()
	// send error to channel
	errorChannel <- fmt.Errorf("System termination signal received")
}
