package user_module

import (
	"codedln/shared/middleware"
	"codedln/shared/mongodb"
	"codedln/shared/redis"
	"codedln/user_module/controller"
	"codedln/user_module/repository"
	"codedln/user_module/service"
	"codedln/util/constant"
	"codedln/util/helpers"
	"codedln/util/types"
	"context"
	"fmt"
	"github.com/go-redis/redis_rate/v10"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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

func TestLogoutUser(t *testing.T) {
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

	collection := db.Collection(constant.UserCollection)
	if collection == nil {
		log.Fatalf("%s does not exist:", constant.UserCollection)
	}

	defer func() {
		_, _ = collection.DeleteMany(context.TODO(), map[string]string{})
	}()

	userRepo := repository.New(collection)
	userService := service.New(userRepo)
	userController := controller.New(userService)

	server := httptest.NewServer(middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
		userController.Logout,
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
		check          bool
		expectedStatus int
	}{
		{
			description:    "Should return a 401 status code with invalid client authorization key",
			endpoint:       fmt.Sprintf("%s", server.URL+"/logout"),
			method:         "DELETE",
			clientKey:      "bad_client_key",
			check:          false,
			expectedStatus: 401,
		},
		{
			description:    "Should return a 200 status code with valid client key",
			endpoint:       fmt.Sprintf("%s", server.URL+"/logout"),
			method:         "DELETE",
			clientKey:      os.Getenv("CLIENT_KEY"),
			check:          true,
			expectedStatus: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &http.Client{}
			req, err := http.NewRequest(test.method, test.endpoint, nil)
			req.Header.Add("Authorization", fmt.Sprintf("%s %s", "Bearer", test.clientKey))
			req.Header.Add("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			if test.check {
				cookie := resp.Header.Get("Set-Cookie")
				accessToken := strings.Split(cookie, ";")
				value := strings.Split(accessToken[0], "=")
				if value[1] != "" {
					t.Errorf("expect access token to be empty")
				}
			}

			if resp.StatusCode != test.expectedStatus {
				t.Errorf("expected %v got %v", test.expectedStatus, resp.StatusCode)
			}

			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected %v got %v", "application/json", resp.Header.Get("Content-Type"))
			}

			if resp.StatusCode >= 400 {
				var body types.ServerResponse
				_ = helpers.JSONDecode(resp.Body, &body)
				fmt.Println(body)
			}
		})
	}
}
