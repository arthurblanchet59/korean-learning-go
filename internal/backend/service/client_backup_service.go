package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/arthurblanchet59/korean-learning-go/internal/backend/domain"
	"github.com/arthurblanchet59/korean-learning-go/internal/backend/repository"
)

const maxClientDocumentSize = 64 * 1024

var ErrInvalidClientBackup = errors.New("invalid client backup")

type ClientBackupService struct {
	backups repository.ClientBackupRepository
	now     func() time.Time
}

func NewClientBackupService(backups repository.ClientBackupRepository) *ClientBackupService {
	return &ClientBackupService{
		backups: backups,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

func (service *ClientBackupService) Backup(ctx context.Context, userID string) (domain.ClientBackup, error) {
	return service.backups.FindClientBackup(ctx, userID)
}

func (service *ClientBackupService) Save(ctx context.Context, userID string, config json.RawMessage, state json.RawMessage) (domain.ClientBackup, error) {
	if err := validateClientDocument("config", config); err != nil {
		return domain.ClientBackup{}, err
	}
	if err := validateClientDocument("state", state); err != nil {
		return domain.ClientBackup{}, err
	}

	backup := domain.ClientBackup{
		UserID:    userID,
		Config:    append(json.RawMessage(nil), config...),
		State:     append(json.RawMessage(nil), state...),
		UpdatedAt: service.now(),
	}
	if err := service.backups.UpsertClientBackup(ctx, backup); err != nil {
		return domain.ClientBackup{}, err
	}
	return backup, nil
}

func validateClientDocument(name string, value json.RawMessage) error {
	if len(value) == 0 || len(value) > maxClientDocumentSize {
		return fmt.Errorf("%w: %s must contain between 1 and %d bytes", ErrInvalidClientBackup, name, maxClientDocumentSize)
	}
	var object map[string]any
	if err := json.Unmarshal(value, &object); err != nil || object == nil {
		return fmt.Errorf("%w: %s must be a valid JSON object", ErrInvalidClientBackup, name)
	}
	return nil
}
