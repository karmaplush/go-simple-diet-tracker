package registration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/karmaplush/simple-diet-tracker/internal/clients/auth/grpc"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/accounts/registration"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/accounts/registration/mocks"

	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/logger/handlers/slogdiscard"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

const (
	successEmail    = "success@example.com"
	notAnEmail      = "randomstring"
	successPassword = "1234test"
	successToken    = "successjwt"
)

func TestRegistrationHandler(t *testing.T) {
	testCases := []struct {
		name                 string
		mockEmail            string
		mockPassword         string
		expectedError        error
		expectedStatusCode   int
		expectedErrorMessage string
		invalidDecoing       bool
	}{
		{
			name:                 "success",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
			expectedError:        nil,
			expectedStatusCode:   http.StatusCreated,
			expectedErrorMessage: "",
			invalidDecoing:       false,
		},
		{
			name:                 "empty email",
			mockEmail:            "",
			mockPassword:         successPassword,
			expectedError:        grpc.ErrInvalidArgument,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "empty password",
			mockEmail:            successEmail,
			mockPassword:         "",
			expectedError:        grpc.ErrInvalidArgument,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "empty email & password",
			mockEmail:            "",
			mockPassword:         "",
			expectedError:        grpc.ErrInvalidArgument,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "not email",
			mockEmail:            notAnEmail,
			mockPassword:         successPassword,
			expectedError:        grpc.ErrInvalidArgument,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "registration: user is exists",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
			expectedError:        grpc.ErrUserExists,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "user is already exists",
			invalidDecoing:       false,
		},
		{
			name:                 "registration: invalid argument",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
			expectedError:        grpc.ErrInvalidArgument,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "invalid credentials",
			invalidDecoing:       false,
		},
		{
			name:                 "registration: unexpected error #1",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
			expectedError:        grpc.ErrGRPCUnexpected,
			expectedStatusCode:   http.StatusInternalServerError,
			expectedErrorMessage: "internal error",
			invalidDecoing:       false,
		},
		{
			name:                 "registration: unexpected error #2",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
			expectedError:        errors.New("some unexpected service error"),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedErrorMessage: "internal error",
			invalidDecoing:       false,
		},
		{
			name:                 "invalid decoded json",
			mockEmail:            successEmail,
			mockPassword:         successPassword,
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

			mockProvider := mocks.NewRegistrationProvider(t)
			mockProvider.On("Registration", mock.AnythingOfType("*context.valueCtx"), tc.mockEmail, tc.mockPassword).
				Return(tc.expectedError).
				Maybe()

			handler := registration.New(slogdiscard.NewDiscardLogger(), mockProvider)

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
				"/accounts/registration",
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
				require.NoError(t, err)
			}
		})
	}
}
