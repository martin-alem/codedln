package user_module_test

import (
	"codedln/shared/middleware"
	"codedln/shared/mongodb"
	"codedln/shared/redis"
	"codedln/user_module/controller"
	"codedln/user_module/model"
	"codedln/user_module/repository"
	"codedln/user_module/service"
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

// New user should be able to create an account if it does not already exist.
// The user should be returned with a cookie set.
// Before Test is being run, a new Google id token must be generated. it's valid for 1 hour
func TestCreateAccount(t *testing.T) {
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
		userController.CreateUser,
		middleware.PayloadValidationMiddleware(model.NewCreateUserSchema),
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
		payload        any
		clientKey      string
		userExist      bool
		authenticate   bool
		expectedStatus int
	}{
		{
			description:    "Should return a 401 status code with invalid client authorization key",
			endpoint:       fmt.Sprintf("%s", server.URL),
			method:         "POST",
			payload:        map[string]string{"signInWith": "google"},
			clientKey:      "bad_client_key",
			userExist:      false,
			authenticate:   false,
			expectedStatus: 401,
		},
		{
			description:    "Should return a 400 status code with no id token",
			endpoint:       fmt.Sprintf("%s", server.URL),
			method:         "POST",
			payload:        map[string]string{"signInWith": "google"},
			clientKey:      os.Getenv("CLIENT_KEY"),
			userExist:      false,
			authenticate:   false,
			expectedStatus: 400,
		},
		{
			description:    "Should return a 400 status code with no sign in option",
			endpoint:       fmt.Sprintf("%s", server.URL),
			method:         "POST",
			payload:        map[string]string{"idToken": "invalidIdToken"},
			clientKey:      os.Getenv("CLIENT_KEY"),
			userExist:      false,
			authenticate:   false,
			expectedStatus: 400,
		},

		{
			description:    "Should return a 201 status code and a valid cookie when valid id token and sign in option are used",
			endpoint:       fmt.Sprintf("%s", server.URL),
			method:         "POST",
			payload:        map[string]string{"idToken": os.Getenv("GOOGLE_ID_TOKEN"), "signInWith": "google"},
			clientKey:      os.Getenv("CLIENT_KEY"),
			userExist:      false,
			authenticate:   true,
			expectedStatus: 201,
		},

		{
			description:    "Should return a 201 status code and a valid cookie when valid id token and sign in option are used if user already exist",
			endpoint:       fmt.Sprintf("%s", server.URL),
			method:         "POST",
			payload:        map[string]string{"idToken": os.Getenv("GOOGLE_ID_TOKEN"), "signInWith": "google"},
			clientKey:      os.Getenv("CLIENT_KEY"),
			userExist:      true,
			authenticate:   true,
			expectedStatus: 201,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if test.userExist {
				_, _ = collection.InsertOne(context.TODO(), map[string]any{"firstname": "Martin", "lastname": "Alemajoh", "email": "alemajohmartin@gmail.com", "verified": true})
			}
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

			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected %v got %v", "application/json", resp.Header.Get("Content-Type"))
			}

			if test.authenticate {
				if resp.Header.Get("Set-Cookie") == "" {
					t.Errorf("expected cookie to be set.")
				}
			}

			if resp.StatusCode >= 200 && resp.StatusCode <= 209 {
				var body struct {
					StatusCode int        `json:"status_code"`
					Data       model.User `json:"data"`
				}
				_ = helpers.JSONDecode(resp.Body, &body)
				if body.Data.Verified != true {
					t.Errorf("expected %v got %v", true, body.Data.Verified)
				}

				if body.Data.FirstName != "Martin" {
					t.Errorf("expected %v got %v", "Martin", body.Data.FirstName)
				}

				if body.Data.LastName != "Alemajoh" {
					t.Errorf("expected %v got %v", "Alemajoh", body.Data.LastName)
				}

				if body.Data.Email != "alemajohmartin@gmail.com" {
					t.Errorf("expected %v got %v", "alemajohmartin@gmail.com", body.Data.Email)
				}
			}
		})
	}

}
