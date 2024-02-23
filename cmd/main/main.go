package main

import (
	"codedln/shared/mongodb"
	"codedln/shared/redis"
	url "codedln/url_module/module"
	user "codedln/user_module/module"
	"codedln/util/helpers"
	"context"
	"errors"
	"github.com/go-redis/redis_rate/v10"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func init() {
	if os.Getenv("ENVIRONMENT") == "development" || os.Getenv("ENVIRONMENT") == "" {
		dotErr := godotenv.Load()
		if dotErr != nil {
			log.Fatalf("Error loading .env file: %v", dotErr)
		}
	}
}

func main() {

	//Initialize mux
	r := mux.NewRouter()

	r.NotFoundHandler = helpers.NotFound()

	//Connect to mongo database
	mClient := mongodb.ConnectToDatabase()

	//Database
	db := mClient.Database(os.Getenv("DATABASE_NAME"))

	//Connect to redis
	rClient := redis.ConnectToRedis()

	//Initialize RateLimiter
	rateLimiter := redis_rate.NewLimiter(rClient)

	//Mount Modules
	user.UserModule(r, rateLimiter, db)
	url.UrlModule(r, rateLimiter, db)

	//Setup Http Server
	server := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 2^20 shifting 1 left by 20 = 1,048,576
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)

	go func() {
		// Setting up a channel to listen for OS signals
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()

		dbErr := mClient.Disconnect(ctx)
		if dbErr != nil {
			log.Fatal("Error disconnecting from database: ", dbErr)
		}

		closeErr := rClient.Close()
		if closeErr != nil {
			log.Fatal("Error disconnecting from redis: ", closeErr)
		}

		if err := server.Shutdown(ctx); err != nil {
			log.Fatal("Server shutdown error: ", err)
		}

		log.Println("Server exiting gracefully")
	}()

	log.Println("Server listening on port 8080")
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("ListenAndServe(): %v", err)
	}
}
