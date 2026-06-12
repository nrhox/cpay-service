package entity

import (
	"time"

	"github.com/nrhox/cpay-service/internal/constants"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID             bson.ObjectID        `bson:"_id,omitempty" json:"_id,omitempty"`
	RoleID         constants.Role       `bson:"role" json:"role,omitempty"`
	FullName       string               `bson:"full_name" json:"full_name,omitempty"`
	Email          string               `bson:"email" json:"email,omitempty"`
	AvatarUrl      string               `bson:"avatar_url" json:"avatar_url,omitempty"`
	Status         constants.UserStatus `bson:"status" json:"status"`
	OAuthProviders []AuthProvider       `bson:"oauth_providers" json:"oauth_providers,omitempty"`
	CreatedAt      time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time            `bson:"updated_at" json:"updated_at"`
}
