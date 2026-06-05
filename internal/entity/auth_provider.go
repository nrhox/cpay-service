package entity

type AuthProvider struct {
	ID       string `bson:"id" json:"id"`
	Provider string `bson:"provider" json:"provider"`
}
