package repository

import (
	"codedln/shared/http_error"
	"codedln/url_module/model"
	"codedln/util/types"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

type UrlRepository interface {
	CreateUrl(ctx context.Context, url model.Url) (*model.Url, error)
	GetUrl(ctx context.Context, filter map[string]any) (*model.Url, error)
	GetUrls(ctx context.Context, searchFilter map[string]any, sortFilter map[string]any, limit int64) (*types.PaginationResult[model.Url], error)
	DeleteUrl(ctx context.Context, filter map[string]any) error
	DeleteUrls(ctx context.Context, filter map[string]any) error
}

type MongoUrlRepository struct {
	collection *mongo.Collection
}

func New(collection *mongo.Collection) UrlRepository {
	return &MongoUrlRepository{
		collection: collection,
	}
}

func (r *MongoUrlRepository) CreateUrl(ctx context.Context, url model.Url) (*model.Url, error) {
	res, err := r.collection.InsertOne(ctx, url)

	if err != nil {
		log.Println(err)
		return nil, http_error.New(http.StatusInternalServerError, "unable to create url")
	}

	var newUrl model.Url

	//Fetch the newly inserted url
	if findErr := r.collection.FindOne(ctx, bson.D{{"_id", res.InsertedID}}).Decode(&newUrl); findErr != nil {
		log.Println(findErr)
		return nil, http_error.New(http.StatusInternalServerError, "unable to find url")
	}

	return &newUrl, nil
}

func (r *MongoUrlRepository) GetUrl(ctx context.Context, filter map[string]any) (*model.Url, error) {

	res := r.collection.FindOne(ctx, filter)
	if errors.Is(res.Err(), mongo.ErrNoDocuments) {
		return nil, nil
	}

	var url model.Url
	if decodeErr := res.Decode(&url); decodeErr != nil {
		log.Println(decodeErr)
		return nil, http_error.New(http.StatusInternalServerError, "unable to get url")
	}

	return &url, nil
}

func (r *MongoUrlRepository) GetUrls(ctx context.Context, searchFilter map[string]any, sortFilter map[string]any, limit int64) (*types.PaginationResult[model.Url], error) {
	// First, calculate the total count of documents matching the searchFilter
	totalCount, err := r.collection.CountDocuments(ctx, searchFilter)
	if err != nil {
		log.Println(err)
		return nil, http_error.New(http.StatusInternalServerError, "unable to count urls")
	}

	// Then, fetch the documents with sorting and limiting
	pipeline := bson.D{
		{Key: "$match", Value: searchFilter},
		{Key: "$sort", Value: sortFilter},
		{Key: "$limit", Value: limit},
	}

	opts := options.Aggregate().SetMaxTime(2 * time.Second)
	cursor, err := r.collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		log.Println(err)
		return nil, http_error.New(http.StatusInternalServerError, "unable to fetch urls")
	}

	var results []model.Url
	if err = cursor.All(ctx, &results); err != nil {
		log.Println(err)
		return nil, http_error.New(http.StatusInternalServerError, "unable to process fetch urls")
	}

	// Return the results along with the total count
	return &types.PaginationResult[model.Url]{
		Data:  results,
		Total: totalCount,
	}, nil
}

func (r *MongoUrlRepository) DeleteUrl(ctx context.Context, filter map[string]any) error {
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println(err)
		return http_error.New(http.StatusInternalServerError, "unable to delete url")
	}

	return nil
}

func (r *MongoUrlRepository) DeleteUrls(ctx context.Context, filter map[string]any) error {
	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Println(err)
		return http_error.New(http.StatusInternalServerError, "unable to delete urls")
	}

	return nil
}
