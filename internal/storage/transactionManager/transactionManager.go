package transactionManager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
	"log"
)

type Transactor struct {
	DB *sql.DB
}

func NewTransactor(DatabaseDSN string) *Transactor {
	log.Println("STORAGE", DatabaseDSN)
	db, err := sql.Open("pgx", DatabaseDSN)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return &Transactor{db}
}

func (t *Transactor) WriteTXSerialize(ctx context.Context, fn func(*sql.Tx) error) error {
	const maxRetries = 3

	for i := 0; i < maxRetries; i++ {
		tx, err := t.DB.BeginTx(ctx, &sql.TxOptions{
			Isolation: sql.LevelSerializable,
		})
		if err != nil {
			return err
		}

		err = fn(tx)
		if err != nil {
			rbErr := tx.Rollback()
			if rbErr != nil {
				return fmt.Errorf("failed to rollback: %v, original error: %w", rbErr, err)
			}
			return err
		}
		if isSerializationError(err) {
			continue
		}

		err = tx.Commit()
		if err != nil {
			if isSerializationError(err) {
				continue
			}
			return fmt.Errorf("failed to commit: %w", err)
		}

		return nil
	}
	return errors.New("faild to send tx")
}

func isSerializationError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "40001"
	}
	return false
}
