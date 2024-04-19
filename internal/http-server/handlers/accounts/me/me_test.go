package me_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/karmaplush/simple-diet-tracker/internal/domain/models"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/accounts/me"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/accounts/me/mocks"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/logger/handlers/slogdiscard"
	"github.com/karmaplush/simple-diet-tracker/internal/services/account"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type contextKey string

const jwtContextKey contextKey = "Token"

func TestMeHandler(t *testing.T) {
	testCases := []struct {
		name                 string
		jwtToken             string
		mockAccount          models.Account
		mockError            error
		expectedCode         int
		expectedErrorMessage string
	}{
		{
			name:                 "success",
			jwtToken:             "good_jwt_token",
			mockAccount:          models.Account{Id: 42, UserId: 42, DailyLimit: 1888},
			mockError:            nil,
			expectedCode:         http.StatusOK,
			expectedErrorMessage: "",
		},
		{
			name:                 "invalid jwt",
			jwtToken:             "bad_jwt_token",
			mockAccount:          models.Account{},
			mockError:            account.ErrInvalidJWT,
			expectedCode:         http.StatusUnauthorized,
			expectedErrorMessage: "incorrect credentials",
		},
		{
			name:                 "account not found",
			jwtToken:             "good_jwt_token",
			mockAccount:          models.Account{},
			mockError:            account.ErrAccountNotFound,
			expectedCode:         http.StatusInternalServerError,
			expectedErrorMessage: "unexpected error",
		},
		{
			name:                 "unexpected error",
			jwtToken:             "maybe_bad_jwt_token_maybe_good_who_knows",
			mockAccount:          models.Account{},
			mockError:            errors.New("🔥🔥🔥🔥🔥"),
			expectedCode:         http.StatusInternalServerError,
			expectedErrorMessage: "unexpected error",
		},
	}

	for _, tc := range testCases {

		tc := tc

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			mockProvider := mocks.NewAccountProvider(t)
			mockProvider.On("GetAccountByContextJWT", mock.AnythingOfType("*context.valueCtx")).
				Return(tc.mockAccount, tc.mockError).
				Once()

			handler := me.New(slogdiscard.NewDiscardLogger(), mockProvider)

			req, err := http.NewRequest(http.MethodGet, "/accounts/me", nil)
			require.NoError(t, err)

			ctx := context.WithValue(req.Context(), jwtContextKey, tc.jwtToken)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			handler(rr, req)

			assert.Equal(t, tc.expectedCode, rr.Code)

			if tc.expectedErrorMessage != "" {
				var errorResponse response.ErrorResponse
				err = json.Unmarshal(rr.Body.Bytes(), &errorResponse)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedErrorMessage, errorResponse.Message)
			} else {
				var acc models.Account
				err = json.Unmarshal(rr.Body.Bytes(), &acc)
				require.NoError(t, err)
				assert.Equal(t, tc.mockAccount, acc)
			}
		})
	}
}
