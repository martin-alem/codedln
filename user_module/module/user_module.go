package module

import (
	"codedln/shared/middleware"
	"codedln/user_module/controller"
	"codedln/user_module/model"
	"codedln/user_module/repository"
	"codedln/user_module/service"
	"codedln/util/constant"
	"github.com/go-redis/redis_rate/v10"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

func UserModule(router *mux.Router, limiter *redis_rate.Limiter, db *mongo.Database) {

	collection := db.Collection(constant.UserCollection)
	if collection == nil {
		log.Fatalf("%s does not exist:", constant.UserCollection)
	}
	userRepo := repository.New(collection)
	userService := service.New(userRepo)
	userController := controller.New(userService)

	userRouter := router.PathPrefix("/user").Subrouter()

	userRouter.HandleFunc("/logout",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			userController.LogOut,
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   10,
				Burst:  5,
				Period: time.Minute * 2,
			}),
		))).Methods("DELETE")

	userRouter.HandleFunc("",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			userController.DeleteUser,
			middleware.AuthenticationMiddleware,
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   100,
				Burst:  50,
				Period: time.Minute * 2,
			}),
		))).Methods("DELETE")

	userRouter.HandleFunc("",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			userController.CreateUser,
			middleware.PayloadValidationMiddleware(model.NewCreateUserSchema),
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   10,
				Burst:  5,
				Period: time.Minute * 2,
			}),
		))).Methods("POST")

	userRouter.HandleFunc("",
		middleware.ExceptionMiddleware(middleware.ChainMiddlewares(
			userController.GetUser,
			middleware.AuthenticationMiddleware,
			middleware.CorsMiddleware,
			middleware.RateLimitMiddleware(limiter, redis_rate.Limit{
				Rate:   100,
				Burst:  50,
				Period: time.Minute * 2,
			}),
		))).Methods("GET")
}
