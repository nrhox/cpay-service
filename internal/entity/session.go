package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Session struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	UserID    bson.ObjectID `bson:"user_id" json:"user_id,omitempty"`
	Token     string        `bson:"token" json:"token,omitempty"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	ExpiredAt time.Time     `bson:"expired_at" json:"expired_at"`
}
