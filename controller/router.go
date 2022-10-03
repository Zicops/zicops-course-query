package controller

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/zicops/contracts/userz"
	"github.com/zicops/zicops-cass-pool/redis"
	"github.com/zicops/zicops-course-query/graph"
	"github.com/zicops/zicops-course-query/graph/generated"
	"github.com/zicops/zicops-course-query/lib/jwt"
	"github.com/zicops/zicops-user-manager/handlers/queries"
)

// CCRouter ... the router for the controller
func CCRouter() (*gin.Engine, error) {
	restRouter := gin.Default()
	// configure cors as needed for FE/BE interactions: For now defaults

	configCors := cors.DefaultConfig()
	configCors.AllowAllOrigins = true
	configCors.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	restRouter.Use(cors.New(configCors))
	// user a middleware to get context values
	restRouter.Use(func(c *gin.Context) {
		currentRequest := c.Request
		incomingToken := jwt.GetToken(currentRequest)
		claimsFromToken, _ := jwt.GetClaims(incomingToken)
		c.Set("zclaims", claimsFromToken)
	})
	restRouter.GET("/healthz", HealthCheckHandler)
	// create group for restRouter
	version1 := restRouter.Group("/api/v1")
	version1.POST("/query", graphqlHandler())
	version1.GET("/playql", playgroundHandler())
	return restRouter, nil
}

func HealthCheckHandler(c *gin.Context) {
	log.Debugf("HealthCheckHandler Method --> %s", c.Request.Method)

	switch c.Request.Method {
	case http.MethodGet:
		GetHealthStatus(c.Writer)
	default:
		err := errors.New("Method not supported")
		ResponseError(c.Writer, http.StatusBadRequest, err)
	}
}

//GetHealthStatus ...
func GetHealthStatus(w http.ResponseWriter) {
	healthStatus := "Super Dentist backend service is healthy"
	response, _ := json.Marshal(healthStatus)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(response); err != nil {
		log.Errorf("GetHealthStatus ... unable to write JSON response: %v", err)
	}
}

// ResponseError ... essentially a single point of sending some error to route back
func ResponseError(w http.ResponseWriter, httpStatusCode int, err error) {
	log.Errorf("Response error %s", err.Error())
	response, _ := json.Marshal(err)
	w.Header().Add("Status", strconv.Itoa(httpStatusCode)+" "+err.Error())
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(httpStatusCode)

	if _, err := w.Write(response); err != nil {
		log.Errorf("ResponseError ... unable to write JSON response: %v", err)
	}
}

func graphqlHandler() gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	// Resolver is in the resolver.go file
	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	return func(c *gin.Context) {
		ctxValue := c.Value("zclaims").(map[string]interface{})
		// set ctxValue to request context
		// get user using email
		emailCalled := ctxValue["email"].(string)
		userIdUsingEmail := base64.URLEncoding.EncodeToString([]byte(emailCalled))
		var userInput userz.User
		// user get query
		redisResult, _ := redis.GetRedisValue(userIdUsingEmail)
		lspIdInt := ctxValue["tenant"]
		lspID := "d8685567-cdae-4ee0-a80e-c187848a760e"
		if lspIdInt != nil && lspIdInt.(string) != "" {
			lspID = lspIdInt.(string)
		}
		ctxValue["lsp_id"] = lspID
		if redisResult != "" {
			err := json.Unmarshal([]byte(redisResult), &userInput)
			if err != nil {
				log.Errorf("Error unmarshalling user from redis %s", err.Error())
			} else {
				ctxValue["role"] = userInput.Role
				redis.SetTTL(userIdUsingEmail, 3600)
			}
		} else {
			user, err := queries.GetUserDetails(c.Request.Context(), []*string{&userInput.ID})
			if err != nil {
				log.Errorf("Error getting user from user manager %s", err.Error())
			}
			if len(user) > 0 {
				ctxValue["role"] = user[0].Role
				userBytes, _ := json.Marshal(user[0])
				redis.SetRedisValue(userIdUsingEmail, string(userBytes))
				redis.SetTTL(userIdUsingEmail, 3600)
			}

		}
		request := c.Request
		requestWithValue := request.WithContext(context.WithValue(request.Context(), "zclaims", ctxValue))
		h.ServeHTTP(c.Writer, requestWithValue)
	}
}

func playgroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
