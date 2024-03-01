package repository

import (
	"codedln/shared/http_error"
	"codedln/url_module/model"
	"codedln/util/types"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"time"
)

type UrlRepository interface {
	CreateUrl(ctx context.Context, url model.Url) (*model.Url, error)
	GetUrl(ctx context.Context, query bson.D) (*model.Url, error)
	GetUrls(ctx context.Context, query string, sort types.DateSort, limit int64, skip int64, userId primitive.ObjectID) (*types.PaginationResult[model.Url], error)
	DeleteUrl(ctx context.Context, urlId primitive.ObjectID, userId primitive.ObjectID) error
	DeleteUrls(ctx context.Context, urlIds []primitive.ObjectID, userId primitive.ObjectID) error
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

func (r *MongoUrlRepository) GetUrl(ctx context.Context, filter bson.D) (*model.Url, error) {

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

func (r *MongoUrlRepository) GetUrls(ctx context.Context, query string, sort types.DateSort, limit int64, skip int64, userId primitive.ObjectID) (*types.PaginationResult[model.Url], error) {

	searchFilter := bson.D{
		{"$and", bson.A{
			bson.D{{"userId", userId}},
			bson.D{{"$or", bson.A{
				bson.D{{"alias", bson.D{{"$regex", query}, {"$options", "i"}}}},
				bson.D{{"originalUrl", bson.D{{"$regex", query}, {"$options", "i"}}}},
			}}},
		}},
	}

	// First, calculate the total count of documents matching the searchFilter
	totalCount, err := r.collection.CountDocuments(ctx, searchFilter)
	if err != nil {
		log.Println(err)
		return nil, http_error.New(http.StatusInternalServerError, "unable to count urls")
	}

	// Then, fetch the documents with sorting and limiting
	pipeline := bson.A{
		bson.D{{Key: "$match", Value: searchFilter}},
		bson.D{{Key: "$sort", Value: bson.D{{"createdAt", sort}}}},
		bson.D{{Key: "$skip", Value: skip}},
		bson.D{{Key: "$limit", Value: limit}},
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

func (r *MongoUrlRepository) DeleteUrl(ctx context.Context, urlId primitive.ObjectID, userId primitive.ObjectID) error {
	filter := bson.D{
		{"$and", bson.A{
			bson.D{{"userId", userId}},
			bson.D{{"_id", urlId}},
		}},
	}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Println(err)
		return http_error.New(http.StatusInternalServerError, "unable to delete url")
	}
	return nil
}

func (r *MongoUrlRepository) DeleteUrls(ctx context.Context, urlIds []primitive.ObjectID, userId primitive.ObjectID) error {
	filter := bson.D{
		{"$and", bson.A{
			bson.D{{"userId", userId}},
			bson.M{"_id": bson.M{"$in": urlIds}},
		}},
	}
	_, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Println(err)
		return http_error.New(http.StatusInternalServerError, "unable to delete urls")
	}
	return nil
}
