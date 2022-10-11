package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"

	"os"

	//"github.com/zicops/zicops-cass-pool/cassandra"
	//"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-cass-pool/cassandra"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/controller"
	"github.com/zicops/zicops-course-query/global"
	cry "github.com/zicops/zicops-course-query/lib/crypto"
)

const defaultPort = "8080"

func main() {
	//os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "zicops-cc.json")
	fmt.Println("Starting zicops course query service")
	ctx, cancel := context.WithCancel(context.Background())
	crySession := cry.New("09afa9f9544a7ff1ae9988f73ba42134")
	global.CTX = ctx
	global.Cancel = cancel
	global.CryptSession = &crySession
	global.Rand = rand.New(rand.NewSource(99))
	fmt.Println("zicops course query initialization complete")
	portFromEnv := os.Getenv("PORT")
	port, err := strconv.Atoi(portFromEnv)

	if err != nil {
		port = 8091
	}
	//test cassandra connection
	_, err1 := cassandra.GetCassSession("coursez")
	_, err2 := cassandra.GetCassSession("qbankz")
	_, err3 := cassandra.GetCassSession("userz")
	if err1 != nil || err2 != nil || err3 != nil {
		fmt.Println("Error connecting to cassandra: %v and %v and %v", err1, err2, err3)
	} else {
		fmt.Println("Cassandra connection successful")
	}
	_, err = redis.Initialize()
	if err != nil {
		fmt.Println("Error connecting to redis: %v", err)
	} else {
		fmt.Println("Redis connection successful")
	}
	bootUPErrors := make(chan error, 1)
	controller.CCBackendController(ctx, port, bootUPErrors)
	err = <-bootUPErrors
	if err != nil {
		fmt.Println("There is an issue starting backend server for course query: %v", err.Error())
		global.WaitGroupServer.Wait()
		os.Exit(1)
	}
	fmt.Println("course query server started successfully.")
}
