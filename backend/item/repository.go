package item

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	coll *mongo.Collection
}

func NewRepository(coll *mongo.Collection) *Repository {
	return &Repository{coll: coll}
}

func (r *Repository) Insert(ctx context.Context, it *Item) (*Item, error) {
	res, err := r.coll.InsertOne(ctx, it)
	if err != nil {
		return nil, err
	}
	it.ID = res.InsertedID.(bson.ObjectID)
	return it, nil
}

func (r *Repository) FindAll(ctx context.Context) ([]Item, error) {
	cur, err := r.coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var items []Item
	if err := cur.All(ctx, &items); err != nil {
		return nil, err
	}
	if items == nil {
		items = []Item{}
	}
	return items, nil
}

func (r *Repository) FindQueue(ctx context.Context) ([]Item, error) {
	cur, err := r.coll.Find(ctx, bson.M{"status": StatusQueued}, options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var items []Item
	if err := cur.All(ctx, &items); err != nil {
		return nil, err
	}
	if items == nil {
		items = []Item{}
	}
	return items, nil
}

func (r *Repository) FindCurrent(ctx context.Context) (*Item, error) {
	var it Item
	err := r.coll.FindOne(ctx, bson.M{"status": StatusDisplaying}, options.FindOne().SetSort(bson.D{{Key: "displayedAt", Value: -1}})).Decode(&it)
	if err != nil {
		return nil, err
	}
	return &it, nil
}

func (r *Repository) FindOldestQueued(ctx context.Context) (*Item, error) {
	var it Item
	err := r.coll.FindOne(ctx, bson.M{"status": StatusQueued}, options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: 1}})).Decode(&it)
	if err != nil {
		return nil, err
	}
	return &it, nil
}

func (r *Repository) FindByID(ctx context.Context, id bson.ObjectID) (*Item, error) {
	var it Item
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&it)
	if err != nil {
		return nil, err
	}
	return &it, nil
}

func (r *Repository) UpdateByID(ctx context.Context, id bson.ObjectID, set bson.M) error {
	_, err := r.coll.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": set})
	return err
}

func (r *Repository) DeleteByID(ctx context.Context, id bson.ObjectID) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
