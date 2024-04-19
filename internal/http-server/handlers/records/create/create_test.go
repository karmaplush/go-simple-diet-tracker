package create_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/records/create"
	"github.com/karmaplush/simple-diet-tracker/internal/http-server/handlers/records/create/mocks"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/logger/handlers/slogdiscard"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/go-playground/assert.v1"
)

var (
	emptyDate     = time.Time{}
	validDate     = time.Now
	emptyValue    = 0
	negativeValue = -600
	validValue    = 500
)

func TestCreateRecordHandler(t *testing.T) {

	testCases := []struct {
		name                 string
		dateRecord           time.Time
		value                int
		expectedError        error
		expectedStatusCode   int
		expectedErrorMessage string
		invalidDecoing       bool
	}{
		{
			name:                 "success",
			dateRecord:           validDate(),
			value:                validValue,
			expectedError:        nil,
			expectedStatusCode:   http.StatusCreated,
			expectedErrorMessage: "",
			invalidDecoing:       false,
		},
		{
			name:                 "empty (default) date record",
			dateRecord:           emptyDate,
			value:                validValue,
			expectedError:        nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "empty (default) value",
			dateRecord:           validDate(),
			value:                emptyValue,
			expectedError:        nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "empty date & value",
			dateRecord:           emptyDate,
			value:                emptyValue,
			expectedError:        nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "negative value",
			dateRecord:           validDate(),
			value:                negativeValue,
			expectedError:        nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "validation failed",
			invalidDecoing:       false,
		},
		{
			name:                 "unexpected service error",
			dateRecord:           validDate(),
			value:                validValue,
			expectedError:        errors.New("some unexpected service layer error was occured"),
			expectedStatusCode:   http.StatusInternalServerError,
			expectedErrorMessage: "unexpected error",
			invalidDecoing:       false,
		},
		{
			name:                 "invalid decoded json",
			dateRecord:           validDate(),
			value:                validValue,
			expectedError:        nil,
			expectedStatusCode:   http.StatusBadRequest,
			expectedErrorMessage: "",
			invalidDecoing:       true,
		},
	}

	for _, tc := range testCases {

		tc := tc

		t.Run(tc.name, func(t *testing.T) {

			t.Parallel()

			mockCreator := mocks.NewRecordCreator(t)
			mockCreator.On(
				"CreateRecordForCurrentUser",
				context.Background(),
				mock.AnythingOfType("time.Time"),
				tc.value,
			).Return(tc.expectedError).Maybe()

			handler := create.New(slogdiscard.NewDiscardLogger(), mockCreator)

			reqBody := fmt.Sprintf(
				`{"value": %d, "dateRecord": "%s"}`,
				tc.value,
				tc.dateRecord.Format("2006-01-02T15:04:05Z"),
			)

			if tc.invalidDecoing {
				reqBody = reqBody[:len(reqBody)-1]
			}

			req, err := http.NewRequest(
				http.MethodPost,
				"/records",
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
