package repository

import (
	"codedln/shared/http_error"
	"codedln/user_module/model"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"net/http"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user model.User) (*model.User, error)
	GetUser(ctx context.Context, filter map[string]any) (*model.User, error)
	DeleteUser(ctx context.Context, userId string) error
}

type MongoUserRepository struct {
	collection *mongo.Collection
}

func New(collection *mongo.Collection) UserRepository {
	return &MongoUserRepository{
		collection: collection,
	}
}

func (r *MongoUserRepository) CreateUser(ctx context.Context, user model.User) (*model.User, error) {
	res, err := r.collection.InsertOne(ctx, user)

	if err != nil {
		log.Println(err)
		return nil, errors.New("unable to create user")
	}

	var newUser model.User

	//Fetch the newly inserted user
	if findErr := r.collection.FindOne(ctx, bson.D{{"_id", res.InsertedID}}).Decode(&newUser); findErr != nil {
		log.Println(findErr)
		return nil, errors.New("unable to find user")
	}

	return &newUser, nil
}

func (r *MongoUserRepository) GetUser(ctx context.Context, filter map[string]any) (*model.User, error) {

	res := r.collection.FindOne(ctx, filter)
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	var user model.User
	if decodeErr := res.Decode(&user); decodeErr != nil {
		return nil, http_error.New(http.StatusInternalServerError, "unable to get user")
	}

	return &user, nil
}

func (r *MongoUserRepository) DeleteUser(ctx context.Context, userId string) error {

	objectId, err := primitive.ObjectIDFromHex(userId)
	filter := map[string]any{"_id": objectId}
	if err != nil {
		return http_error.New(http.StatusBadRequest, "trying to delete user with invalid id")
	}

	opts := options.Session().SetDefaultReadConcern(readconcern.Majority())
	sess, err := r.collection.Database().Client().StartSession(opts)
	if err != nil {
		return http_error.New(http.StatusInternalServerError, "error starting transaction")
	}
	defer sess.EndSession(ctx)

	txnOpts := options.Transaction().SetReadPreference(readpref.PrimaryPreferred())
	_, err = sess.WithTransaction(ctx, func(txnCtx mongo.SessionContext) (interface{}, error) {
		//Delete all data associated with user

		//Delete user
		_, deleteErr := r.collection.DeleteOne(ctx, filter)
		if deleteErr != nil {
			return http_error.New(http.StatusInternalServerError, "error deleting user"), nil
		}
		return nil, nil
	}, txnOpts)

	if err != nil {
		return http_error.New(http.StatusBadRequest, "trying to delete user with invalid id")
	}

	return nil
}
