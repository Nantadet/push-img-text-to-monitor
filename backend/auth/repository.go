package auth

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Repository struct {
	coll *mongo.Collection
}

func NewRepository(coll *mongo.Collection) *Repository {
	return &Repository{coll: coll}
}

func (r *Repository) Insert(ctx context.Context, a *Auth) (*Auth, error) {
	res, err := r.coll.InsertOne(ctx, a)
	if err != nil {
		return nil, err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		a.ID = id
	}
	return a, nil
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*Auth, error) {
	var a Auth
	if err := r.coll.FindOne(ctx, bson.M{"username": username}).Decode(&a); err != nil {
		return nil, err
	}
	return &a, nil
}
