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
	"errors"
	"fmt"
	"github.com/go-redis/redis_rate/v10"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func TestDeleteUser(t *testing.T) {
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

	r, _ := collection.InsertOne(context.TODO(), map[string]any{"firstname": "Martin", "lastname": "Alemajoh", "email": "alemajohmartin@gmail.com", "verified": true})

	defer func() {
		_, _ = collection.DeleteMany(context.TODO(), map[string]string{})
	}()

	userRepo := repository.New(collection)
	userService := service.New(userRepo)
	userController := controller.New(userService)

	server := httptest.NewServer(middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
		userController.DeleteUser,
		middleware.AuthenticationMiddleware,
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
		setCookie      bool
		clientKey      string
		jwt            string
		check          bool
		expectedStatus int
	}{
		{
			description:    "Should return a 401 status code with invalid client authorization key",
			endpoint:       fmt.Sprintf("%s", server.URL),
			method:         "DELETE",
			setCookie:      false,
			clientKey:      "bad_client_key",
			jwt:            "",
			check:          false,
			expectedStatus: 401,
		},
		{
			description:    "Should return a 401 status code with no cookie set",
			endpoint:       fmt.Sprintf("%s", server.URL),
			method:         "DELETE",
			setCookie:      false,
			clientKey:      os.Getenv("CLIENT_KEY"),
			jwt:            "",
			check:          false,
			expectedStatus: 401,
		},
		{
			description:    "Should return a 401 status code with invalid jwt",
			endpoint:       fmt.Sprintf("%s", server.URL),
			method:         "DELETE",
			setCookie:      true,
			clientKey:      os.Getenv("CLIENT_KEY"),
			jwt:            "invalid_token",
			check:          false,
			expectedStatus: 401,
		},
		{
			description:    "Should return a 401 status code with expired jwt",
			endpoint:       fmt.Sprintf("%s", server.URL),
			method:         "DELETE",
			setCookie:      true,
			clientKey:      os.Getenv("CLIENT_KEY"),
			jwt:            helpers.CreateJWT(types.AuthUser{UserId: r.InsertedID.(primitive.ObjectID).Hex()}, -1),
			check:          false,
			expectedStatus: 401,
		},
		{
			description:    "Should return a 200 status code with valid jwt",
			endpoint:       fmt.Sprintf("%s", server.URL),
			method:         "DELETE",
			setCookie:      true,
			clientKey:      os.Getenv("CLIENT_KEY"),
			jwt:            helpers.CreateJWT(types.AuthUser{UserId: r.InsertedID.(primitive.ObjectID).Hex()}, constant.AccessTokenTTL),
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
			if test.setCookie {
				cookie := &http.Cookie{
					Name:  constant.JwtCookieName,
					Value: test.jwt,
				}
				req.AddCookie(cookie)
			}
			resp, err := client.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			if test.check {
				s := collection.FindOne(context.TODO(), map[string]any{"_id": r.InsertedID})
				if !errors.Is(s.Err(), mongo.ErrNoDocuments) {
					t.Errorf("expected document to have been deleted")
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
