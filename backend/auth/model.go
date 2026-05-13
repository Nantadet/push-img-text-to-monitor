package auth

import "go.mongodb.org/mongo-driver/v2/bson"

type Auth struct {
	ID           bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string        `bson:"username" json:"username"`
	PasswordHash string        `bson:"passwordHash" json:"-"`
}
