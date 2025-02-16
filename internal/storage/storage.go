package storage

import (
	"avito/internal/models"
	"avito/internal/storage/transactionManager"
	"context"
	"database/sql"
	"errors"
)

type Storage interface {
	GetUser(ctx context.Context, username string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error

	GetInfo(ctx context.Context, uuid string) (*models.Info, error)

	Send(ctx context.Context, uuid string, toUser string, amount int) error

	Buy(ctx context.Context, uuid string, item string) error

	GetUuidByUsername(ctx context.Context, username string) (string, error)
}

type DataBase struct {
	Tm *transactionManager.Transactor
}

func NewStorage(DatabaseDSN string) *DataBase {
	return &DataBase{Tm: transactionManager.NewTransactor(DatabaseDSN)}
}

func (db *DataBase) GetUser(ctx context.Context, username string) (*models.User, error) {
	row := db.Tm.DB.QueryRowContext(ctx, `
		SELECT id, username, password_hash, balance, created_at
		FROM users
		WHERE username = $1
	`, username)

	user := &models.User{}
	err := row.Scan(&user.UUID, &user.Username, &user.PasswordHash, &user.Balance, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (db *DataBase) CreateUser(ctx context.Context, user *models.User) error {
	_, err := db.Tm.DB.ExecContext(ctx, `
		INSERT INTO users (id, username, password_hash, balance)
		VALUES ($1, $2, $3, $4)
	`, user.UUID, user.Username, user.PasswordHash, user.Balance)
	return err
}

func (db *DataBase) GetInfo(ctx context.Context, uuid string) (*models.Info, error) {
	info := &models.Info{}

	// get user Balance
	var balance int
	err := db.Tm.DB.QueryRowContext(ctx, `
		SELECT balance
		  FROM users
		 WHERE id = $1
	`, uuid).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, err
	}

	// get transactions history
	info.Coins = balance
	received, err := db.getReceivedTransactions(ctx, uuid)
	if err != nil {
		return nil, err
	}
	sent, err := db.getSentTransactions(ctx, uuid)
	if err != nil {
		return nil, err
	}
	info.CoinsHistory = models.History{
		Received: received,
		Sent:     sent,
	}
	inventory, err := db.getUserInventory(ctx, uuid)
	if err != nil {
		return nil, err
	}
	info.Inventory = inventory

	return info, nil
}

func (db *DataBase) Send(ctx context.Context, uuid string, toUser string, amount int) error {
	return db.Tm.WriteTXSerialize(ctx, func(tx *sql.Tx) error {
		balance, err := db.getBalanceTx(ctx, tx, uuid)
		if err != nil {
			return err
		}
		if balance-amount < 0 {
			return ErrNotEnoughBalance
		}
		var receiverUuid string
		err = tx.QueryRowContext(ctx, `
		SELECT id
		  FROM users
		 WHERE username = $1
	`, toUser).Scan(&receiverUuid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrUserNotFound
			}
			return err
		}
		if uuid == receiverUuid {
			return ErrSendingToYourself
		}
		err = db.updateBalance(ctx, tx, uuid, balance-amount)
		if err != nil {
			return err
		}
		balance, err = db.getBalanceTx(ctx, tx, receiverUuid)
		if err != nil {
			return err
		}
		err = db.updateBalance(ctx, tx, receiverUuid, balance+amount)
		if err != nil {
			return err
		}
		err = db.createTransaction(ctx, tx, uuid, receiverUuid, amount)
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *DataBase) Buy(ctx context.Context, uuid string, item string) error {
	return db.Tm.WriteTXSerialize(ctx, func(tx *sql.Tx) error {
		balance, err := db.getBalanceTx(ctx, tx, uuid)
		if err != nil {
			return err
		}
		var itemPrice int
		err = tx.QueryRowContext(ctx, "SELECT price FROM merchandise WHERE name = $1", item).Scan(&itemPrice)
		if err != nil {
			return err
		}
		if balance-itemPrice < 0 {
			return ErrNotEnoughBalance
		}
		err = db.updateBalance(ctx, tx, uuid, balance-itemPrice)
		if err != nil {
			return err
		}
		err = db.updateInventory(ctx, tx, uuid, item)
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *DataBase) getItemPrice(ctx context.Context, item string) (int, error) {
	var price int
	err := db.Tm.DB.QueryRowContext(ctx, `
		SELECT price
		  FROM merchandise
		 WHERE name = $1
	`, item).Scan(&price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrItemNotFound
		}
		return 0, err
	}
	return price, nil
}

func (db *DataBase) GetUuidByUsername(ctx context.Context, username string) (string, error) {
	var userUUID string
	err := db.Tm.DB.QueryRowContext(ctx, `
		SELECT id
		  FROM users
		 WHERE username = $1
	`, username).Scan(&userUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	return userUUID, nil
}

func (db *DataBase) getUsernameByUuid(ctx context.Context, uuid string) (string, error) {
	var username string
	err := db.Tm.DB.QueryRowContext(ctx, `
		SELECT username
		  FROM users
		 WHERE id = $1
	`, uuid).Scan(&username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrUserNotFound
		}
		return "", err
	}
	return username, nil
}

func (db *DataBase) getBalanceTx(ctx context.Context, tx *sql.Tx, uuid string) (int, error) {
	var balance int
	err := tx.QueryRowContext(ctx, `
		SELECT balance
		  FROM users
		 WHERE id = $1
	`, uuid).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrUserNotFound
		}
		return 0, err
	}
	return balance, nil
}

func (db *DataBase) updateBalance(ctx context.Context, tx *sql.Tx, uuid string, updatedBalance int) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE users
		   SET balance = $1
		 WHERE id = $2
	`, updatedBalance, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}
	return err
}

func (db *DataBase) createTransaction(ctx context.Context, tx *sql.Tx, senderId, receiverId string, amount int) error {
	_, err := tx.ExecContext(ctx, `
        INSERT INTO transactions (sender_id, receiver_id, amount)
        VALUES ($1, $2, $3)
    `, senderId, receiverId, amount)
	if err != nil {
		return err
	}

	return nil
}

func (db *DataBase) updateInventory(ctx context.Context, tx *sql.Tx, uuid string, item string) error {
	var merchID int

	// get merchID
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM merchandise
		 WHERE name = $1
	`, item).Scan(&merchID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrItemNotFound
		}
		return err
	}

	// check is this user have this merchId
	var existingID string
	var existingQty int
	err = tx.QueryRowContext(ctx, `
		SELECT id, quantity
		  FROM user_inventory
		 WHERE user_id = $1
		   AND merchandise_id = $2
	`, uuid, merchID).Scan(&existingID, &existingQty)

	// create new if user don't have such merch
	if errors.Is(err, sql.ErrNoRows) {
		_, insertErr := tx.ExecContext(ctx, `
			INSERT INTO user_inventory (user_id, merchandise_id, quantity)
			VALUES ($1, $2, 1)
		`, uuid, merchID)
		return insertErr
	} else if err != nil {
		return err
	}

	_, updateErr := tx.ExecContext(ctx, `
		UPDATE user_inventory
		   SET quantity = $1
		 WHERE id = $2
	`, existingQty+1, existingID)

	return updateErr
}

func (db *DataBase) getReceivedTransactions(ctx context.Context, uuid string) ([]models.Transaction, error) {
	rows, err := db.Tm.DB.QueryContext(ctx, `
		SELECT sender.username AS from_user,
		       receiver.username AS to_user,
		       t.amount
		  FROM transactions t
		  JOIN users sender   ON t.sender_id = sender.id
		  JOIN users receiver ON t.receiver_id = receiver.id
		 WHERE t.receiver_id = $1
		 ORDER BY t.created_at DESC
	`, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Transaction
	for rows.Next() {
		var tr models.Transaction
		if scanErr := rows.Scan(&tr.FromUser, &tr.ToUser, &tr.Amount); scanErr != nil {
			return nil, scanErr
		}
		result = append(result, tr)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return result, nil
}

func (db *DataBase) getSentTransactions(ctx context.Context, uuid string) ([]models.Transaction, error) {
	rows, err := db.Tm.DB.QueryContext(ctx, `
		SELECT sender.username AS from_user,
		       receiver.username AS to_user,
		       t.amount
		  FROM transactions t
		  JOIN users sender   ON t.sender_id = sender.id
		  JOIN users receiver ON t.receiver_id = receiver.id
		 WHERE t.sender_id = $1
		 ORDER BY t.created_at DESC
	`, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Transaction
	for rows.Next() {
		var tr models.Transaction
		if scanErr := rows.Scan(&tr.FromUser, &tr.ToUser, &tr.Amount); scanErr != nil {
			return nil, scanErr
		}
		result = append(result, tr)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return result, nil
}

func (db *DataBase) getUserInventory(ctx context.Context, uuid string) ([]models.Item, error) {
	rows, err := db.Tm.DB.QueryContext(ctx, `
		SELECT m.name, ui.quantity
		  FROM user_inventory ui
		  JOIN merchandise m ON ui.merchandise_id = m.id
		 WHERE ui.user_id = $1
	`, uuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inventory []models.Item
	for rows.Next() {
		var item models.Item
		if scanErr := rows.Scan(&item.Type, &item.Quantity); scanErr != nil {
			return nil, scanErr
		}
		inventory = append(inventory, item)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, rowsErr
	}
	return inventory, nil
}
