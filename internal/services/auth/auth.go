package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/karmaplush/simple-diet-tracker/internal/clients/auth/grpc"
	"github.com/karmaplush/simple-diet-tracker/internal/domain/models"
	"github.com/karmaplush/simple-diet-tracker/internal/storage"
)

type Auth struct {
	log                  *slog.Logger
	loginProvider        LoginProvider
	registrationProvider RegistrationProvider
	accountProvider      AccountProvider
	accountCreator       AccountCreator
}

type LoginProvider interface {
	Login(
		ctx context.Context,
		email string,
		password string,
	) (token string, userId int64, err error)
}

type RegistrationProvider interface {
	Registration(ctx context.Context, email string, password string) (userId int64, err error)
}

type AccountProvider interface {
	AccountByUserId(ctx context.Context, userId int64) (models.Account, error)
}

type AccountCreator interface {
	SaveAccount(ctx context.Context, userId int64) (int64, error)
}

func New(
	log *slog.Logger,
	loginProvider LoginProvider,
	registrationProvider RegistrationProvider,
	accountProvider AccountProvider,
	accountCreator AccountCreator,
) *Auth {
	return &Auth{
		log:                  log,
		loginProvider:        loginProvider,
		registrationProvider: registrationProvider,
		accountProvider:      accountProvider,
		accountCreator:       accountCreator,
	}
}

func (a *Auth) Login(
	ctx context.Context,
	email string,
	password string,
) (string, error) {
	const op = "services.auth.Login"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	token, userId, err := a.loginProvider.Login(ctx, email, password)

	if err != nil {
		if errors.Is(err, grpc.ErrUserNotFound) || errors.Is(err, grpc.ErrInvalidArgument) {
			log.Info("invalid credentials")
		} else {
			log.Error("failed to login", slog.String("err", err.Error()))
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	_, err = a.accountProvider.AccountByUserId(ctx, userId)
	if err != nil {

		if errors.Is(err, storage.ErrAccountNotFound) {

			_, err = a.accountCreator.SaveAccount(ctx, userId)
			if err != nil {
				log.Error("failed to save account", slog.String("err", err.Error()))
				return "", fmt.Errorf("%s: %w", op, err)
			}

			return token, nil
		}

		log.Error("failed to get account by user id", slog.String("err", err.Error()))
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil

}

func (a *Auth) Registration(
	ctx context.Context,
	email string,
	password string,
) error {
	const op = "services.auth.Registration"

	log := a.log.With(slog.String("op", op), slog.String("email", email))

	userId, err := a.registrationProvider.Registration(ctx, email, password)
	if err != nil {
		if errors.Is(err, grpc.ErrInvalidArgument) || errors.Is(err, grpc.ErrUserExists) {
			log.Info("invalid credentials")
		} else {
			log.Error("failed to register", slog.String("err", err.Error()))
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = a.accountCreator.SaveAccount(ctx, userId)
	if err != nil {
		log.Error("failed to save account", slog.String("err", err.Error()))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}
