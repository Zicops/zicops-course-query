package main

import (
	"context"
	"fmt"
	"math/rand"
	"os/signal"
	"strconv"
	"syscall"

	"os"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	//"github.com/zicops/zicops-cass-pool/cassandra"
	//"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/controller"
	"github.com/zicops/zicops-course-query/global"
	cry "github.com/zicops/zicops-course-query/lib/crypto"
)

func main() {
	//os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "zicops-cc.json")
	log.Infof("Starting zicops course query service")
	ctx, cancel := context.WithCancel(context.Background())
	crySession := cry.New("09afa9f9544a7ff1ae9988f73ba42134")
	global.CTX = ctx
	global.Cancel = cancel
	global.CryptSession = &crySession
	global.Rand = rand.New(rand.NewSource(99))
	log.Infof("zicops course query initialization complete")
	portFromEnv := os.Getenv("PORT")
	port, err := strconv.Atoi(portFromEnv)

	if err != nil {
		port = 8091
	}
	gin.SetMode(gin.ReleaseMode)
	_, err1 := cassandra.GetCassSession("coursez")
	if err1 != nil {
		log.Errorf("Error connecting to cassandra: %v", err1)
	} else {
		log.Infof("Cassandra connection successful")
	}
	_, err2 := cassandra.GetCassSession("qbankz")
	if err2 != nil {
		log.Errorf("Error connecting to cassandra: %v", err2)
	} else {
		log.Infof("Cassandra connection successful")
	}
	_, err3 := cassandra.GetCassSession("userz")
	if err3 != nil {
		log.Errorf("Error connecting to cassandra: %v", err3)
	} else {
		log.Infof("Cassandra connection successful")
	}
	_, err4 := redis.Initialize()
	if err4 != nil {
		log.Errorf("Error connecting to redis: %v", err4)
	} else {
		log.Infof("Redis connection successful")
	}

	bootUPErrors := make(chan error, 1)
	go monitorSystem(cancel, bootUPErrors)
	controller.CCBackendController(ctx, port, bootUPErrors)
	err = <-bootUPErrors
	if err != nil {
		log.Errorf("there is an issue starting backend server for course query: %v", err.Error())
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
	errorChannel <- fmt.Errorf("system termination signal received")
}
