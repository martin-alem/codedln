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
	"net/http"
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
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   1000,
				Burst:  500,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods(http.MethodGet)

	urlRouter.HandleFunc("/check_alias",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.CheckAliasExistence,
			middleware.PayloadValidationMiddleware(model.NewCheckAliasSchema),
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   10,
				Burst:  5,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods(http.MethodPost)

	urlRouter.HandleFunc("/guest",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.CreateUrl,
			middleware.PayloadValidationMiddleware(model.NewCreateUrlSchema),
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   10,
				Burst:  5,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods(http.MethodPost)

	urlRouter.HandleFunc("/get_url/{urlId}",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.GetUrl,
			middleware.AuthenticationMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   100,
				Burst:  50,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods(http.MethodGet)

	urlRouter.HandleFunc("/delete_url/{urlId}",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.DeleteUrl,
			middleware.AuthenticationMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   100,
				Burst:  50,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods(http.MethodDelete)

	urlRouter.HandleFunc("/create_url",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.CreateUrl,
			middleware.AuthenticationMiddleware,
			middleware.PayloadValidationMiddleware(model.NewCreateUrlSchema),
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   10,
				Burst:  5,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods(http.MethodPost)

	urlRouter.HandleFunc("/get_urls",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.GetUrls,
			middleware.AuthenticationMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   1000,
				Burst:  500,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods(http.MethodGet)

	urlRouter.HandleFunc("/delete_urls",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			urlController.DeleteUrls,
			middleware.AuthenticationMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   10,
				Burst:  50,
				Period: time.Minute * 2,
			}),
			middleware.ValidateAPIKeyMiddleware,
		))).Methods(http.MethodDelete)

}
