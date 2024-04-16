package login_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/karmaplush/simple-diet-tracker/internal/clients/auth/grpc"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/accounts/login"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/accounts/login/mocks"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/logger/handlers/slogdiscard"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

func TestLoginHandler(t *testing.T) {
	testCases := []struct {
		name                 string
		mockEmail            string
		mockPassword         string
		requestBody          string
		expectedToken        string
		expectedError        error
		expectedCode         int
		expectedErrorMessage string
	}{
		{
			name:                 "success",
			mockEmail:            "success@example.com",
			mockPassword:         "1234test",
			requestBody:          `{"email": "success@example.com", "password": "1234test"}`,
			expectedToken:        "mockjwt",
			expectedError:        nil,
			expectedCode:         http.StatusOK,
			expectedErrorMessage: "",
		},
		{
			name:                 "empty email",
			mockEmail:            "",
			mockPassword:         "1234test",
			requestBody:          `{"email": "", "password": "1234test"}`,
			expectedToken:        "",
			expectedError:        grpc.ErrInvalidArgument,
			expectedCode:         http.StatusUnauthorized,
			expectedErrorMessage: "validation failed",
		},
		{
			name:                 "empty password",
			mockEmail:            "error@example.com",
			mockPassword:         "",
			requestBody:          `{"email": "success@example.com", "password": ""}`,
			expectedToken:        "",
			expectedError:        grpc.ErrInvalidArgument,
			expectedCode:         http.StatusUnauthorized,
			expectedErrorMessage: "validation failed",
		},
		{
			name:                 "empty email & password",
			mockEmail:            "",
			mockPassword:         "",
			requestBody:          `{"email": "", "password": ""}`,
			expectedToken:        "",
			expectedError:        grpc.ErrInvalidArgument,
			expectedCode:         http.StatusUnauthorized,
			expectedErrorMessage: "validation failed",
		},
		{
			name:                 "not email",
			mockEmail:            "somerandom12345string",
			mockPassword:         "1234test",
			requestBody:          `{"email": "somerandom12345string", "password": "1234test"}`,
			expectedToken:        "",
			expectedError:        grpc.ErrInvalidArgument,
			expectedCode:         http.StatusUnauthorized,
			expectedErrorMessage: "validation failed",
		},
		{
			name:                 "invalid json",
			mockEmail:            "invalid@example.com",
			mockPassword:         "1234test",
			requestBody:          `{"email": ", "password": "}`,
			expectedToken:        "",
			expectedError:        nil,
			expectedCode:         http.StatusBadRequest,
			expectedErrorMessage: "",
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

			handler := login.New(slogdiscard.NewDiscardLogger(), mockProvider)

			req, err := http.NewRequest(
				http.MethodPost,
				"/accounts/login",
				bytes.NewReader([]byte(tc.requestBody)),
			)
			require.NoError(t, err)

			responseRecorder := httptest.NewRecorder()
			handler(responseRecorder, req)

			assert.Equal(t, tc.expectedCode, responseRecorder.Code)

			if tc.expectedErrorMessage != "" {
				var resp response.ErrorResponse
				err = json.Unmarshal(responseRecorder.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedErrorMessage, resp.Message)
			} else {
				var resp login.Response
				err = json.Unmarshal(responseRecorder.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, tc.expectedToken, resp.Token)
			}
		})
	}
}
