package push

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, p *Push) (*Push, error) {
	return s.repo.Insert(ctx, p)
}

func (s *Service) List(ctx context.Context) ([]Push, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) Get(ctx context.Context, id bson.ObjectID) (*Push, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id bson.ObjectID, set bson.M) error {
	if len(set) == 0 {
		return nil
	}
	return s.repo.Update(ctx, id, set)
}

func (s *Service) Delete(ctx context.Context, id bson.ObjectID) error {
	return s.repo.Delete(ctx, id)
}
