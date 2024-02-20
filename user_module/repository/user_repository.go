package repository

import (
	"codedln/shared/http_error"
	"codedln/user_module/model"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"net/http"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user model.User) (*model.User, error)
	GetUser(ctx context.Context, filter map[string]string) (*model.User, error)
	DeleteUser(ctx, context, userId string) *model.User
}

type MongoRepository struct {
	collection *mongo.Collection
}

func New(collection *mongo.Collection) UserRepository {
	return &MongoRepository{
		collection: collection,
	}
}

func (r *MongoRepository) CreateUser(ctx context.Context, user model.User) (*model.User, error) {
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

func (r *MongoRepository) GetUser(ctx context.Context, filter map[string]string) (*model.User, error) {

	res := r.collection.FindOne(ctx, filter)
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	var user model.User
	if decodeErr := res.Decode(&user); decodeErr != nil {
		return nil, http_error.New(http.StatusInternalServerError, "unable to get user")
	}

	return &user, errors.New("user exist")
}

func (r *MongoRepository) DeleteUser(ctx, context, userId string) *model.User {
	//TODO implement me
	panic("implement me")
}
