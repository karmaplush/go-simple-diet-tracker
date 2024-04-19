package delete_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	deleteHandler "github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/records/delete"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/records/delete/mocks"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/logger/handlers/slogdiscard"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

const (
	correctRecordIdParam   = "10"
	incorrectRecordIdParam = "invalid"
	invalidRecordIdParam   = "-888"
	invalidRecordIdParam2  = "0"
)

func TestDeleteRecordHandler(t *testing.T) {

	testCases := []struct {
		name                 string
		recordIdPathParam    string
		expectedError        error
		expectedStatusCode   int
		expectedErrorMessage string
	}{
		{
			name:                 "success",
			recordIdPathParam:    correctRecordIdParam,
			expectedError:        nil,
			expectedStatusCode:   http.StatusNoContent,
			expectedErrorMessage: "",
		},
		{
			name:                 "incorrect recordId param",
			recordIdPathParam:    incorrectRecordIdParam,
			expectedError:        nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "invalid record id",
		},
		{
			name:                 "invalid recordId param",
			recordIdPathParam:    invalidRecordIdParam,
			expectedError:        nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
		},
		{
			name:                 "invalid recordId param #2",
			recordIdPathParam:    invalidRecordIdParam2,
			expectedError:        nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
		},
		{
			name:                 "unexpected service error",
			recordIdPathParam:    "10",
			expectedError:        errors.New("some unexpected service layer error was occured"),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedErrorMessage: "unexpected error",
		},
	}

	for _, tc := range testCases {

		tc := tc

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			mockRemover := mocks.NewRecordRemover(t)
			mockRemover.On(
				"DeleteRecordForCurrentUser",
				mock.Anything,
				mock.AnythingOfType("int64"),
			).Return(tc.expectedError).Maybe()

			router := chi.NewRouter()
			router.Use(middleware.URLFormat)

			handler := deleteHandler.New(slogdiscard.NewDiscardLogger(), mockRemover)
			router.Delete("/records/{recordId}", handler)

			url := fmt.Sprintf("/records/%s", tc.recordIdPathParam)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			responseRecorder := httptest.NewRecorder()

			router.ServeHTTP(responseRecorder, req)

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
