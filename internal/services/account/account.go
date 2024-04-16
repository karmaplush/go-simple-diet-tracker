package account

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-chi/jwtauth"
	"github.com/karmaplush/simple-diet-tracker/internal/domain/models"
	"github.com/karmaplush/simple-diet-tracker/internal/storage"
)

type Account struct {
	log             *slog.Logger
	accountProvider AccountProvider
	accountSaver    AccountSaver
}

type AccountProvider interface {
	AccountById(ctx context.Context, accountId int64) (account models.Account, err error)
	AccountByUserId(ctx context.Context, userId int64) (account models.Account, err error)
}

type AccountSaver interface {
	SaveAccount(ctx context.Context, userId int64) (uid int64, err error)
}

var (
	ErrAccountNotFound = errors.New("account not found")
	ErrAccountExists   = errors.New("account exists")
	ErrInvalidJWT      = errors.New("invalid jwt")
)

func New(
	log *slog.Logger,
	accountProvider AccountProvider,
	accountSaver AccountSaver,
) *Account {
	return &Account{
		log:             log,
		accountProvider: accountProvider,
		accountSaver:    accountSaver,
	}
}

func (a *Account) GetAccountById(
	ctx context.Context,
	accountId int64,
) (models.Account, error) {
	const op = "services.account.GetAccountById"

	log := a.log.With(slog.String("op", op), slog.Int64("account_id", accountId))

	account, err := a.accountProvider.AccountById(ctx, accountId)

	if err != nil {
		if errors.Is(err, storage.ErrAccountNotFound) {
			log.Info("account not found")

			return models.Account{}, fmt.Errorf("%s: %w", op, ErrAccountNotFound)
		}

		log.Error("failed to get account", slog.String("err", err.Error()))
		return models.Account{}, fmt.Errorf("%s: %w", op, err)
	}

	return account, nil

}

func (a *Account) GetAccountByUserId(
	ctx context.Context,
	userId int64,
) (models.Account, error) {
	const op = "services.account.GetAccountByUserId"

	log := a.log.With(slog.String("op", op), slog.Int64("user_id", userId))

	account, err := a.accountProvider.AccountByUserId(ctx, userId)

	if err != nil {
		if errors.Is(err, storage.ErrAccountNotFound) {
			log.Info("account not found")
			return models.Account{}, fmt.Errorf("%s: %w", op, ErrAccountNotFound)
		}

		log.Error("failed to get account", slog.String("err", err.Error()))
		return models.Account{}, fmt.Errorf("%s: %w", op, err)
	}

	return account, nil

}

func (a *Account) GetAccountByContextJWT(ctx context.Context) (models.Account, error) {
	const op = "services.account.GetAccountByContextJWT"

	log := a.log.With(slog.String("op", op))

	// Get JWT claims from JWT context
	_, claims, err := jwtauth.FromContext(ctx)
	if err != nil {
		log.Error("jwt claims extracting from context failed", slog.String("err", err.Error()))
		return models.Account{}, fmt.Errorf("%s: %w", op, ErrInvalidJWT)
	}

	// Get uid claim data (UserID Int64)
	uid, exists := claims["uid"]
	if !exists {
		log.Error("missing uid claim in JWT")
		return models.Account{}, fmt.Errorf("%s: %w", op, ErrInvalidJWT)
	}

	// Check if the type of the 'uid' is correct
	_, ok := uid.(float64)
	if !ok {
		log.Error("incorrect uid claim type")
		return models.Account{}, fmt.Errorf("%s: %w", op, ErrInvalidJWT)
	}

	userId := int64(uid.(float64))

	return a.GetAccountByUserId(ctx, userId)
}
