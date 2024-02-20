package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	FirstName string             `bson:"firstname" json:"firstname"`
	LastName  string             `bson:"lastname" json:"lastname"`
	Email     string             `bson:"email" json:"email"`
	Picture   string             `bson:"picture" json:"picture"`
	Verified  bool               `bson:"verified" json:"verified"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
