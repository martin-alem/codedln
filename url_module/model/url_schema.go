package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Url struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserId      primitive.ObjectID `bson:"userId,omitempty" json:"userId"`
	OriginalUrl string             `bson:"originalUrl" json:"originalUrl"`
	Alias       string             `bson:"alias" json:"alias"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
