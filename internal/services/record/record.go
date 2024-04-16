package record

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/karmaplush/simple-diet-tracker/internal/domain/models"
)

type Record struct {
	log             *slog.Logger
	recordProvider  RecordProvider
	recordSaver     RecordSaver
	recordRemover   RecordRemover
	accountProvider AccountProvider
}

type RecordProvider interface {
	RecordById(ctx context.Context, recordId int64) (account models.Record, err error)
	RecordsByUserId(
		ctx context.Context,
		userId int64,
		date time.Time,
	) (records []models.Record, err error)
}

type RecordSaver interface {
	SaveRecord(
		ctx context.Context,
		accountId int64,
		value int,
		dateRecord time.Time,
	) (int64, error)
}

type RecordRemover interface {
	DeleteRecord(ctx context.Context, accountId int64, recordId int64) error
}

type AccountProvider interface {
	GetAccountByContextJWT(ctx context.Context) (models.Account, error)
}

var (
	ErrRecordNotFound = errors.New("record not found")
)

func New(
	log *slog.Logger,
	recordProvider RecordProvider,
	recordSaver RecordSaver,
	recordRemover RecordRemover,
	accountProvider AccountProvider,
) *Record {
	return &Record{
		log:             log,
		recordProvider:  recordProvider,
		recordSaver:     recordSaver,
		recordRemover:   recordRemover,
		accountProvider: accountProvider,
	}
}

func (r *Record) GetRecordsForCurrentUser(
	ctx context.Context,
	date time.Time,
) ([]models.Record, error) {
	const op = "services.record.GetRecordsForCurrentUser"

	log := r.log.With(slog.String("op", op))

	acc, err := r.accountProvider.GetAccountByContextJWT(ctx)
	if err != nil {
		log.Error("can not get records - incorrect token")
		return []models.Record{}, fmt.Errorf("%s: %w", op, err)
	}

	records, err := r.recordProvider.RecordsByUserId(ctx, acc.UserId, date)
	if err != nil {
		log.Error("can not get records")
		return []models.Record{}, fmt.Errorf("%s: %w", op, err)
	}

	if len(records) == 0 {
		records = []models.Record{}
	}

	return records, nil
}

func (r *Record) CreateRecordForCurrentUser(
	ctx context.Context,
	dateRecord time.Time,
	value int,
) error {
	const op = "services.record.CreateRecordForCurrentUser"

	log := r.log.With(slog.String("op", op))

	acc, err := r.accountProvider.GetAccountByContextJWT(ctx)
	if err != nil {
		log.Error("can not create record - incorrect token")
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = r.recordSaver.SaveRecord(ctx, acc.Id, value, dateRecord)
	if err != nil {
		log.Error("failed to save record", slog.String("err", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}

func (r *Record) DeleteRecordForCurrentUser(ctx context.Context, recordId int64) error {
	const op = "services.record.DeleteRecordForCurrentUser"

	log := r.log.With(slog.String("op", op))

	acc, err := r.accountProvider.GetAccountByContextJWT(ctx)
	if err != nil {
		log.Error("can not delete record - incorrect token")
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := r.recordRemover.DeleteRecord(ctx, acc.Id, recordId); err != nil {
		log.Error("failed to delete record", slog.String("err", err.Error()))
	}

	return nil
}
