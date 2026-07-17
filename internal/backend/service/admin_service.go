package service

import (
	"context"

	"github.com/arthurblanchet59/korean-learning-go/internal/backend/repository"
)

type AdminService struct {
	admin repository.AdminRepository
}

func NewAdminService(admin repository.AdminRepository) *AdminService {
	return &AdminService{admin: admin}
}

func (service *AdminService) ResetDatabase(ctx context.Context) (repository.ResetResult, error) {
	return service.admin.Reset(ctx)
}
