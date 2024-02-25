package url_module_test

import (
	"codedln/shared/middleware"
	"codedln/shared/mongodb"
	"codedln/shared/redis"
	"codedln/url_module/controller"
	"codedln/url_module/model"
	"codedln/url_module/repository"
	"codedln/url_module/service"
	"codedln/util/constant"
	"codedln/util/helpers"
	"context"
	"fmt"
	"github.com/go-redis/redis_rate/v10"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "" {
		wd, _ := os.Getwd()
		dotErr := godotenv.Load(filepath.Join(wd, "../../../", ".env"))
		if dotErr != nil {
			log.Fatalf("Error loading .env file: %v", dotErr)
		}
	}
}

func TestCheckAlias(t *testing.T) {
	t.Parallel()
	client := mongodb.ConnectToDatabase()

	defer func() {
		_ = client.Disconnect(context.TODO())
	}()

	//Connect to redis
	rClient := redis.ConnectToRedis()
	defer func() {
		_ = client.Disconnect(context.TODO())
	}()

	//Initialize RateLimiter
	rateLimiter := redis_rate.NewLimiter(rClient)

	db := client.Database("codedln_test_database")

	userCollection := db.Collection(constant.UserCollection)
	if userCollection == nil {
		log.Fatalf("%s does not exist:", constant.UserCollection)
	}

	urlCollection := db.Collection(constant.UrlCollection)
	if urlCollection == nil {
		log.Fatalf("%s does not exist:", constant.UrlCollection)
	}

	user, _ := userCollection.InsertOne(context.TODO(), map[string]any{"firstname": "Martin", "lastname": "Alemajoh", "email": "alemajohmartin@gmail.com", "verified": true})
	_, _ = urlCollection.InsertOne(context.TODO(), map[string]any{"userId": user.InsertedID, "originalUrl": "https://google.com", "alias": "Huixyk"})
	_, _ = urlCollection.InsertOne(context.TODO(), map[string]any{"originalUrl": "https://facebook.com", "alias": "Uinxhj"})

	defer func() {
		_, _ = userCollection.DeleteMany(context.TODO(), map[string]string{})
		_, _ = urlCollection.DeleteMany(context.TODO(), map[string]string{})
	}()

	urlRepo := repository.New(urlCollection)
	urlService := service.New(urlRepo)
	urlController := controller.New(urlService)

	server := httptest.NewServer(middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
		urlController.CheckAliasExistence,
		middleware.PayloadValidationMiddleware(model.NewCheckAliasSchema),
		middleware.CorsMiddleware,
		middleware.RateLimitMiddleware(rateLimiter, redis_rate.Limit{
			Rate:   10,
			Burst:  5,
			Period: time.Minute * 2,
		}),
		middleware.ValidateAPIKeyMiddleware,
	)))

	defer server.Close()

	tests := []struct {
		description    string
		endpoint       string
		method         string
		clientKey      string
		payload        any
		expectedStatus int
	}{
		{
			description:    "Should return a 401 status code with invalid client authorization key",
			endpoint:       fmt.Sprintf("%s", server.URL+"/check_alias"),
			method:         "POST",
			clientKey:      "bad_client_key",
			payload:        map[string]any{"invalid_key": "hikxys"},
			expectedStatus: 401,
		},
		{
			description:    "Should return a 400 status code with no payload",
			endpoint:       fmt.Sprintf("%s", server.URL+"/check_alias"),
			method:         "POST",
			clientKey:      os.Getenv("CLIENT_KEY"),
			payload:        map[string]any{"invalid_key": "hikxys"},
			expectedStatus: 400,
		},

		{
			description:    "Should return a 400 status code with invalid payload",
			endpoint:       fmt.Sprintf("%s", server.URL+"/check_alias"),
			method:         "POST",
			clientKey:      os.Getenv("CLIENT_KEY"),
			payload:        map[string]any{"invalid_key": "hikxys"},
			expectedStatus: 400,
		},

		{
			description:    "Should return a 200 status code with alias that does not exist",
			endpoint:       fmt.Sprintf("%s", server.URL+"/check_alias"),
			method:         "POST",
			clientKey:      os.Getenv("CLIENT_KEY"),
			payload:        map[string]any{"alias": "hikxys"},
			expectedStatus: 200,
		},

		{
			description:    "Should return a 400 status code with alias that exist",
			endpoint:       fmt.Sprintf("%s", server.URL+"/check_alias"),
			method:         "POST",
			clientKey:      os.Getenv("CLIENT_KEY"),
			payload:        map[string]any{"alias": "Huixyk"},
			expectedStatus: 400,
		},

		{
			description:    "Should return a 400 status code with alias that exist",
			endpoint:       fmt.Sprintf("%s", server.URL+"/check_alias"),
			method:         "POST",
			clientKey:      os.Getenv("CLIENT_KEY"),
			payload:        map[string]any{"alias": "Uinxhj"},
			expectedStatus: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &http.Client{}
			payload, err := helpers.AnyTypeToReader(test.payload)
			req, err := http.NewRequest(test.method, test.endpoint, payload)
			req.Header.Add("Authorization", fmt.Sprintf("%s %s", "Bearer", test.clientKey))
			req.Header.Add("Content-Type", "application/json")
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			if resp.StatusCode != test.expectedStatus {
				t.Errorf("expected %v got %v", test.expectedStatus, resp.StatusCode)
			}

		})
	}
}
