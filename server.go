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

	bootUPErrors := make(chan error, 1)
	go monitorSystem(cancel, bootUPErrors)
	go checkAndInitCassandraSession()
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

func checkAndInitCassandraSession() {
	// get user session every 1 minute
	// if session is nil then create new session
	//test cassandra connection
	_, err := redis.Initialize()
	if err != nil {
		log.Errorf("Error connecting to redis: %v", err)
	} else {
		log.Infof("Redis connection successful")
	}
	_, err1 := cassandra.GetCassSession("coursez")
	_, err2 := cassandra.GetCassSession("qbankz")
	if err1 != nil || err2 != nil {
		log.Errorf("Error connecting to cassandra: %v and %v ", err1, err2)
	} else {
		log.Infof("Cassandra connection successful")
	}
}
