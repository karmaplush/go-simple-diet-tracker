package list_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/karmaplush/simple-diet-tracker/internal/domain/models"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/records/list"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/records/list/mocks"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/logger/handlers/slogdiscard"
	"github.com/karmaplush/simple-diet-tracker/internal/services/account"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

var (
	// Truncate for removing monotonic part from Time struct
	mockDate    time.Time       = time.Now().Truncate(0)
	mockRecords []models.Record = []models.Record{
		{
			Id:          1,
			AccountId:   1,
			Value:       500,
			DateRecord:  mockDate,
			DateCreated: mockDate,
		},
	}
)

func TestRecordsListHandler(t *testing.T) {
	testCases := []struct {
		name                 string
		dateQueryParam       string
		mockRecords          []models.Record
		expectedError        error
		expectedStatusCode   int
		expectedErrorMessage string
	}{
		{
			name:                 "success",
			dateQueryParam:       "",
			mockRecords:          mockRecords,
			expectedError:        nil,
			expectedStatusCode:   http.StatusOK,
			expectedErrorMessage: "",
		},
		{
			name:                 "success #2",
			dateQueryParam:       "2024-04-19",
			mockRecords:          mockRecords,
			expectedError:        nil,
			expectedStatusCode:   http.StatusOK,
			expectedErrorMessage: "",
		},
		{
			name:                 "invalid query param value",
			dateQueryParam:       "invalid",
			mockRecords:          mockRecords,
			expectedError:        nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "invalid date format (YYYY-MM-DD format expected)",
		},
		{
			name:                 "service layer: invalid jwt",
			dateQueryParam:       "",
			mockRecords:          mockRecords,
			expectedError:        account.ErrInvalidJWT,
			expectedStatusCode:   http.StatusUnauthorized,
			expectedErrorMessage: "invalid credentials",
		},
		{
			name:                 "service layer: unexpected error",
			dateQueryParam:       "",
			mockRecords:          mockRecords,
			expectedError:        errors.New("some unexpected service layer error was occured"),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedErrorMessage: "unexpected error",
		},
	}

	for _, tc := range testCases {

		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			mockProvider := mocks.NewRecordProvider(t)
			mockProvider.On(
				"GetRecordsForCurrentUser",
				mock.Anything,
				mock.AnythingOfType("time.Time"),
			).Return(tc.mockRecords, tc.expectedError).Maybe()

			handler := list.New(slogdiscard.NewDiscardLogger(), mockProvider)

			url := fmt.Sprintf("/records/?date=%s", tc.dateQueryParam)
			req, err := http.NewRequest(http.MethodGet, url, nil)
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
				var records []models.Record
				err = json.Unmarshal(responseRecorder.Body.Bytes(), &records)
				require.NoError(t, err)
				assert.Equal(t, tc.mockRecords[0].Id, records[0].Id)
				assert.Equal(t, tc.mockRecords[0].AccountId, records[0].AccountId)
				assert.Equal(t, tc.mockRecords[0].Value, records[0].Value)
				assert.Equal(t, tc.mockRecords[0].DateRecord, records[0].DateRecord)
				assert.Equal(t, tc.mockRecords[0].DateCreated, records[0].DateCreated)
			}

		})
	}
}
