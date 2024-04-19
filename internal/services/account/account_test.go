package account_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/karmaplush/simple-diet-tracker/internal/domain/models"
	"github.com/karmaplush/simple-diet-tracker/internal/services/account"
	"github.com/karmaplush/simple-diet-tracker/internal/services/account/mocks"
	"github.com/karmaplush/simple-diet-tracker/internal/storage"
	"github.com/stretchr/testify/mock"
	"gopkg.in/go-playground/assert.v1"
)

func TestAccount_GetAccountById(t *testing.T) {

	const methodOp = "services.account.GetAccountById"

	testCases := []struct {
		name          string
		mockAccountId int64
		mockAccount   models.Account
		mockError     error
		expectedError error
	}{
		{
			name:          "success",
			mockAccountId: 10,
			mockAccount:   models.Account{Id: 10, UserId: 1, DailyLimit: 2000},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "account not found",
			mockAccountId: 888,
			mockAccount:   models.Account{},
			mockError:     storage.ErrAccountNotFound,
			expectedError: account.ErrAccountNotFound,
		},
		{
			name:          "unexpected error",
			mockAccountId: 10,
			mockAccount:   models.Account{},
			mockError:     errors.New("unexpected error"),
			expectedError: fmt.Errorf("%s: %w", methodOp, errors.New("unexpected error")),
		},
	}

	for _, tc := range testCases {

		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			mockProvider := mocks.NewAccountProvider(t)
			mockProvider.On("AccountById", mock.Anything, tc.mockAccountId).
				Return(tc.mockAccount, tc.mockError)

			service := account.New(slog.Default(), mockProvider, nil)

			result, err := service.GetAccountById(context.Background(), tc.mockAccountId)

			assert.Equal(t, tc.mockAccount, result)
			assert.IsEqual(err, tc.expectedError)

		})
	}
}
