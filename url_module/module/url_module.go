package module

import (
	"codedln/shared/middleware"
	"codedln/url_module/controller"
	"codedln/url_module/model"
	"codedln/url_module/repository"
	"codedln/url_module/service"
	"codedln/util/constant"
	"github.com/go-redis/redis_rate/v10"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

func UrlModule(router *mux.Router, limiter *redis_rate.Limiter, db *mongo.Database) {
	collection := db.Collection(constant.UrlCollection)
	if collection == nil {
		log.Fatalf("%s does not exist:", constant.UrlCollection)
	}

	urlRepo := repository.New(collection)
	urlService := service.New(urlRepo)
	urlController := controller.New(urlService)

	urlRouter := router.PathPrefix("/url").Subrouter()

	urlRouter.HandleFunc("/redirect",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.Redirect,
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   1000,
				Burst:  500,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods("GET")

	urlRouter.HandleFunc("/check_alias",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.CheckAliasExistence,
			middleware.PayloadValidationMiddleware(model.NewCheckAliasSchema),
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   10,
				Burst:  5,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods("POST")

	urlRouter.HandleFunc("/guest",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.CreateUrl,
			middleware.PayloadValidationMiddleware(model.NewCreateUrlSchema),
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   10,
				Burst:  5,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods("POST")

	urlRouter.HandleFunc("/{urlId}",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.GetUrl,
			middleware.AuthenticationMiddleware,
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   100,
				Burst:  50,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods("GET")

	urlRouter.HandleFunc("/{urlId}",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.DeleteUrl,
			middleware.AuthenticationMiddleware,
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   100,
				Burst:  50,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods("DELETE")

	urlRouter.HandleFunc("",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.CreateUrl,
			middleware.AuthenticationMiddleware,
			middleware.PayloadValidationMiddleware(model.NewCreateUrlSchema),
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   10,
				Burst:  5,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods("POST")

	urlRouter.HandleFunc("",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.GetUrls,
			middleware.AuthenticationMiddleware,
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   1000,
				Burst:  500,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods("GET")

	urlRouter.HandleFunc("",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.DeleteUrls,
			middleware.AuthenticationMiddleware,
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   1000,
				Burst:  500,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods("DELETE")

}
