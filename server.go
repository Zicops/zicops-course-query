package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os/signal"
	"strconv"
	"syscall"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/penglongli/gin-metrics/ginmetrics"
	log "github.com/sirupsen/logrus"

	"github.com/zicops/contracts/coursez"
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
	// get global Monitor object
	m := ginmetrics.GetMonitor()

	// +optional set metric path, default /debug/metrics
	m.SetMetricPath("/metrics")
	// +optional set slow time, default 5s
	m.SetSlowTime(10)
	// +optional set request duration, default {0.1, 0.3, 1.2, 5, 10}
	// used to p95, p99
	m.SetDuration([]float64{0.1, 0.3, 1.2, 5, 10})

	// set middleware for gin
	r := gin.Default()
	m.Use(r)
	gin.SetMode(gin.ReleaseMode)

	bootUPErrors := make(chan error, 1)
	go monitorSystem(cancel, bootUPErrors)
	go checkAndInitCassandraSession()
	controller.CCBackendController(ctx, port, bootUPErrors, r)
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
	dbCourses := make([]coursez.Course, 0)
	c1 := coursez.Course{
		ID: "1",
	}
	c2 := coursez.Course{
		ID: "2",
	}
	dbCourses = append(dbCourses, c1)
	dbCourses = append(dbCourses, c2)
	redisBytes, err := json.Marshal(dbCourses)
	if err != nil {
		log.Errorf("Error marshalling redis value: %v", err)
	}
	err = redis.SetRedisValue(context.Background(), "test", string(redisBytes))
	if err != nil {
		log.Errorf("Error setting redis value: %v", err)
	}
	err = redis.SetTTL(context.Background(), "test", 60)
	if err != nil {
		log.Errorf("Failed to set redis value: %v", err.Error())
	}
	vaulue, err := redis.GetRedisValue(context.Background(), "test")
	if err != nil {
		log.Errorf("Error getting redis value: %v", err)
	}
	if vaulue == "" {
		log.Errorf("Redis value is empty")
	} else {
		err = json.Unmarshal([]byte(vaulue), &dbCourses)
		if err != nil {
			log.Errorf("Error unmarshalling redis value: %v", err)
		}
	}
	ctx := context.Background()
	cassPool := cassandra.GetCassandraPoolInstance()
	global.CassPool = cassPool
	_, err1 := global.CassPool.GetSession(ctx, "coursez")
	_, err2 := global.CassPool.GetSession(ctx, "qbankz")
	if err1 != nil || err2 != nil {
		log.Errorf("Error connecting to cassandra: %v and %v ", err1, err2)
	} else {
		log.Infof("Cassandra connection successful")
	}
}
