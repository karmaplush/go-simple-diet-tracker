package list

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/karmaplush/simple-diet-tracker/internal/domain/models"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
	"github.com/karmaplush/simple-diet-tracker/internal/services/account"
)

//go:generate go run github.com/vektra/mockery/v2@v2.42.2 --name=RecordProvider
type RecordProvider interface {
	GetRecordsForCurrentUser(ctx context.Context, date time.Time) ([]models.Record, error)
}

const (
	expectedQueryDateFormat = "2006-01-02"
)

func New(
	log *slog.Logger,
	recordProvider RecordProvider,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.records.list.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		// Default value for records date (today)
		recordsDate := time.Now().Truncate(24 * time.Hour)

		dateQueryParam := r.URL.Query().Get("date")

		if dateQueryParam != "" {
			parsedDate, err := time.Parse(expectedQueryDateFormat, dateQueryParam)
			if err != nil {
				render.Status(r, http.StatusBadRequest)
				render.JSON(w, r,
					response.ErrorMessage("invalid date format (YYYY-MM-DD format expected)"),
				)
				return
			}

			recordsDate = parsedDate.Truncate(24 * time.Hour)
		}

		records, err := recordProvider.GetRecordsForCurrentUser(r.Context(), recordsDate)

		if err != nil {

			if errors.Is(err, account.ErrInvalidJWT) {
				render.Status(r, http.StatusUnauthorized)
				render.JSON(w, r, response.ErrorMessage("invalid credentials"))
				return
			}

			log.Error("unexpected error", slog.String("err", err.Error()))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, response.ErrorMessage("unexpected error"))
			return
		}

		render.JSON(w, r, records)
	}
}
