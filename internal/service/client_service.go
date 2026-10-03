package service

import (
	"context"
	"fmt"

	"ims-finance/internal/model"
	"ims-finance/internal/repository"
)

type ClientService struct {
	clientRepo *repository.ClientRepository
}

func NewClientService(repo *repository.ClientRepository) *ClientService {
	return &ClientService{clientRepo: repo}
}

func (s *ClientService) GetAll(ctx context.Context) ([]model.Client, error) {
	return s.clientRepo.GetAll(ctx)
}

func (s *ClientService) GetByID(ctx context.Context, id int64) (*model.Client, error) {
	c, err := s.clientRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("client not found")
	}
	return c, nil
}

func (s *ClientService) Create(ctx context.Context, req *model.CreateClientRequest) (*model.Client, error) {
	id, err := s.clientRepo.Insert(ctx, req)
	if err != nil {
		return nil, err
	}
	return s.clientRepo.GetByID(ctx, id)
}

func (s *ClientService) Update(ctx context.Context, id int64, req *model.UpdateClientRequest) error {
	return s.clientRepo.Update(ctx, id, req)
}

func (s *ClientService) Delete(ctx context.Context, id int64) error {
	return s.clientRepo.Delete(ctx, id)
}
