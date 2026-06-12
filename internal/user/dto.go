package user

type UserInfo struct {
	FullName  string `bson:"full_name" json:"full_name,omitempty"`
	Email     string `bson:"email" json:"email,omitempty"`
	AvatarUrl string `bson:"avatar_url" json:"avatar_url,omitempty"`
}
