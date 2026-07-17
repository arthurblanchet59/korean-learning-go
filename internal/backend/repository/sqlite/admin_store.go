package sqlite

import (
	"context"

	"github.com/arthurblanchet59/korean-learning-go/internal/backend/repository"
)

func (store *Store) Reset(ctx context.Context) (repository.ResetResult, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return repository.ResetResult{}, err
	}
	defer rollback(tx)

	result := repository.ResetResult{}
	var resetErr error

	if result.DeletedReviews, resetErr = execDelete(ctx, tx, `DELETE FROM reviews`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if result.DeletedCards, resetErr = execDelete(ctx, tx, `DELETE FROM cards`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if result.DeletedDecks, resetErr = execDelete(ctx, tx, `DELETE FROM decks`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if result.DeletedJournal, resetErr = execDelete(ctx, tx, `DELETE FROM journal_entries`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if result.DeletedProgress, resetErr = execDelete(ctx, tx, `DELETE FROM lesson_progress`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if _, resetErr = execDelete(ctx, tx, `DELETE FROM user_seed_versions`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}
	if result.DeletedUsers, resetErr = execDelete(ctx, tx, `DELETE FROM users WHERE is_admin = 0`); resetErr != nil {
		return repository.ResetResult{}, resetErr
	}

	if err := tx.Commit(); err != nil {
		return repository.ResetResult{}, err
	}

	return result, nil
}
