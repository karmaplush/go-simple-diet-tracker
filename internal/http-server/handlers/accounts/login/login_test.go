package login_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/karmaplush/simple-diet-tracker/internal/clients/auth/grpc"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/accounts/login"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/accounts/login/mocks"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

const (
	successEmail    = "success@example.com"
	notAnEmail      = "randomstring"
	successPassword = "1234test"
	successToken    = "successjwt"
)

func TestLoginHandler(t *testing.T) {
	testCases := []struct {
		name                 string
		mockEmail            string
		mockPassword         string
		expectedToken        string
		expectedError        error
		expectedStatusCode   int
		expectedErrorMessage string
		invalidDecoing       bool
	}{
		{
			name:                 "success",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
			expectedToken:        successToken,
			expectedError:        nil,
			expectedStatusCode:   http.StatusOK,
			expectedErrorMessage: "",
			invalidDecoing:       false,
		},
		{
			name:                 "empty email",
			mockEmail:            "",
			mockPassword:         successPassword,
			expectedToken:        "",
			expectedError:        grpc.ErrInvalidArgument,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "empty password",
			mockEmail:            successEmail,
			mockPassword:         "",
			expectedToken:        "",
			expectedError:        grpc.ErrInvalidArgument,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "empty email & password",
			mockEmail:            "",
			mockPassword:         "",
			expectedToken:        "",
			expectedError:        grpc.ErrInvalidArgument,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "not email",
			mockEmail:            notAnEmail,
			mockPassword:         successPassword,
			expectedToken:        "",
			expectedError:        grpc.ErrInvalidArgument,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "login: invalid argument",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
			expectedToken:        successToken,
			expectedError:        grpc.ErrInvalidArgument,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedErrorMessage: "invalid credentials",
			invalidDecoing:       false,
		},
		{
			name:                 "login: user not found",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
			expectedToken:        successToken,
			expectedError:        grpc.ErrUserNotFound,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedErrorMessage: "invalid credentials",
			invalidDecoing:       false,
		},
		{
			name:                 "login: unexpected error #1",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
			expectedToken:        successToken,
			expectedError:        grpc.ErrGRPCUnexpected,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedErrorMessage: "internal error",
			invalidDecoing:       false,
		},
		{
			name:                 "login: unexpected error #2",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
			expectedToken:        successToken,
			expectedError:        errors.New("some unexpected service error"),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedErrorMessage: "internal error",
			invalidDecoing:       false,
		},
		{
			name:                 "invalid decoded json",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
			expectedToken:        successToken,
			expectedError:        nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "invalid request",
			invalidDecoing:       true,
		},
	}

	for _, tc := range testCases {

		tc := tc

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			mockProvider := mocks.NewLoginProvider(t)
			mockProvider.On("Login", context.Background(), tc.mockEmail, tc.mockPassword).
				Return(tc.expectedToken, tc.expectedError).
				Maybe()

			handler := login.New(slog.Default(), mockProvider)

			reqBody := fmt.Sprintf(
				`{"email": "%s", "password": "%s"}`,
				tc.mockEmail,
				tc.mockPassword,
			)

			if tc.invalidDecoing {
				reqBody = reqBody[:len(reqBody)-1]
			}

			req, err := http.NewRequest(
				http.MethodPost,
				"/accounts/login",
				bytes.NewReader([]byte(reqBody)),
			)

			require.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler(responseRecorder, req)

			assert.Equal(t, tc.expectedStatusCode, responseRecorder.Code)

			if tc.expectedErrorMessage != "" {
				var errorResponse response.ErrorResponse
				err = json.Unmarshal(responseRecorder.Body.Bytes(), &errorResponse)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedErrorMessage, errorResponse.Message)
			} else {
				var resp login.Response
				err = json.Unmarshal(responseRecorder.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedToken, resp.Token)
			}
		})
	}
}
