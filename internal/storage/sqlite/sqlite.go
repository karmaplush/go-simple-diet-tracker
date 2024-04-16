package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/karmaplush/simple-diet-tracker/internal/domain/models"
	"github.com/karmaplush/simple-diet-tracker/internal/storage"
	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

const (
	defaultDailyLimit = 2000
)

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveAccount(
	ctx context.Context,
	userId int64,
) (int64, error) {
	const op = "storage.sqlite.SaveAccount"

	stmt, err := s.db.Prepare("INSERT INTO accounts(user_id, daily_limit) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.ExecContext(ctx, userId, defaultDailyLimit)
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrAccountExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) AccountById(ctx context.Context, accountID int64) (models.Account, error) {
	const op = "storage.sqlite.AccountById"

	stmt, err := s.db.Prepare("SELECT id, user_id, daily_limit FROM accounts WHERE id = ?")
	if err != nil {
		return models.Account{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, accountID)

	var account models.Account

	err = row.Scan(&account.Id, &account.UserId, &account.DailyLimit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Account{}, fmt.Errorf("%s: %w", op, storage.ErrAccountNotFound)
		}
		return models.Account{}, fmt.Errorf("%s: %w", op, err)
	}

	return account, nil
}

func (s *Storage) AccountByUserId(ctx context.Context, userId int64) (models.Account, error) {
	const op = "storage.sqlite.AccountByUserId"

	stmt, err := s.db.Prepare("SELECT id, user_id, daily_limit FROM accounts WHERE user_id = ?")
	if err != nil {
		return models.Account{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, userId)

	var account models.Account

	err = row.Scan(&account.Id, &account.UserId, &account.DailyLimit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Account{}, fmt.Errorf("%s: %w", op, storage.ErrAccountNotFound)
		}
		return models.Account{}, fmt.Errorf("%s: %w", op, err)
	}

	return account, nil
}

func (s *Storage) SaveRecord(
	ctx context.Context,
	accountId int64,
	value int,
	dateRecord time.Time,
) (int64, error) {
	const op = "storage.sqlite.SaveRecord"

	stmt, err := s.db.Prepare(
		"INSERT INTO records(account_id, value, date_record, date_created) VALUES (?, ?, ?, ?)",
	)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, accountId, value, dateRecord, time.Now())
	if err != nil {
		var sqliteErr sqlite3.Error

		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrAccountExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) RecordById(
	ctx context.Context,
	recordId int64,
) (models.Record, error) {
	const op = "storage.sqlite.RecordById"

	stmt, err := s.db.Prepare(
		"SELECT id, account_id, value, date_record, date_created FROM records WHERE record_id = ?",
	)
	if err != nil {
		return models.Record{}, fmt.Errorf("%s: %w", op, err)
	}

	row := stmt.QueryRowContext(ctx, recordId)

	var record models.Record

	err = row.Scan(
		&record.Id,
		&record.AccountId,
		&record.Value,
		&record.DateRecord,
		&record.DateCreated,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Record{}, fmt.Errorf("%s: %w", op, storage.ErrRecordNotFound)
		}
		return models.Record{}, fmt.Errorf("%s: %w", op, err)
	}

	return record, nil

}

func (s *Storage) RecordsByUserId(
	ctx context.Context,
	userId int64,
	date time.Time,
) ([]models.Record, error) {
	const op = "storage.sqlite.RecordsByUserId"

	stmt, err := s.db.Prepare(`
		SELECT records.id, account_id, value, date_record, date_created
		FROM records
		JOIN accounts ON records.account_id = accounts.id
		WHERE accounts.user_id = ? AND date(records.date_record) = date(?)
		ORDER BY date_created DESC
	`,
	)
	if err != nil {
		return []models.Record{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userId, date)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var records []models.Record

	for rows.Next() {
		var record models.Record

		err := rows.Scan(
			&record.Id,
			&record.AccountId,
			&record.Value,
			&record.DateRecord,
			&record.DateCreated,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		records = append(records, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return records, nil
}

func (s *Storage) DeleteRecord(ctx context.Context, accountId int64, recordId int64) error {
	const op = "storage.sqlite.DeleteRecord"

	stmt, err := s.db.Prepare("DELETE FROM records WHERE account_id = ? AND id = ?")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, accountId, recordId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
