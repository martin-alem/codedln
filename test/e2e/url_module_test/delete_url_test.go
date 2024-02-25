package url_module_test

import (
	"codedln/shared/middleware"
	"codedln/shared/mongodb"
	"codedln/shared/redis"
	"codedln/url_module/controller"
	"codedln/url_module/repository"
	"codedln/url_module/service"
	"codedln/util/constant"
	"codedln/util/helpers"
	"codedln/util/types"
	"context"
	"fmt"
	"github.com/go-redis/redis_rate/v10"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func TestDeleteUrl(t *testing.T) {
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

	server := httptest.NewServer(
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.DeleteUrl,
			middleware.AuthenticationMiddleware,
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(rateLimiter, redis_rate.Limit{
				Rate:   100,
				Burst:  50,
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
		expectedStatus int
	}{
		{
			description:    "Should return a 401 status code with invalid client authorization key",
			endpoint:       fmt.Sprintf("%s", server.URL+"/delete_url/1233455"),
			method:         "DELETE",
			clientKey:      "bad_client_key",
			expectedStatus: 401,
		},
		{
			description:    "Should return a 401 status code with no cookie set",
			endpoint:       fmt.Sprintf("%s", server.URL+"/delete_url/23578695"),
			method:         "DELETE",
			setCookie:      false,
			clientKey:      os.Getenv("CLIENT_KEY"),
			jwt:            "",
			expectedStatus: 401,
		},
		{
			description:    "Should return a 401 status code with invalid jwt",
			endpoint:       fmt.Sprintf("%s", server.URL+"/delete_url/23578695"),
			method:         "DELETE",
			setCookie:      true,
			clientKey:      os.Getenv("CLIENT_KEY"),
			jwt:            "invalid_token",
			expectedStatus: 401,
		},
		{
			description:    "Should return a 400 status code with invalid url id",
			endpoint:       fmt.Sprintf("%s", server.URL+"/delete_url/23578695"),
			method:         "DELETE",
			setCookie:      true,
			clientKey:      os.Getenv("CLIENT_KEY"),
			jwt:            helpers.CreateJWT(types.AuthUser{UserId: user.InsertedID.(primitive.ObjectID).Hex()}, constant.AccessTokenTTL),
			expectedStatus: 400,
		},
		{
			description:    "Should return a 401 status code with expired jwt",
			endpoint:       fmt.Sprintf("%s", server.URL+"/delete_url/12233445"),
			method:         "DELETE",
			setCookie:      true,
			clientKey:      os.Getenv("CLIENT_KEY"),
			jwt:            helpers.CreateJWT(types.AuthUser{UserId: user.InsertedID.(primitive.ObjectID).Hex()}, -1),
			expectedStatus: 401,
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

			if resp.StatusCode != test.expectedStatus {
				t.Errorf("expected %v got %v", test.expectedStatus, resp.StatusCode)
			}

			if resp.StatusCode >= 400 {
				var body types.ServerResponse
				_ = helpers.JSONDecode(resp.Body, &body)
				fmt.Println(body)
			}

		})
	}
}
