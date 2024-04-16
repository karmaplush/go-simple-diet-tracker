package delete

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/karmaplush/simple-diet-tracker/internal/lib/api/response"
)

type RecordRemover interface {
	DeleteRecordForCurrentUser(
		ctx context.Context,
		recordId int64,
	) error
}

type PathParams struct {
	RecordId int64 `validate:"required,gte=1"`
}

func New(
	log *slog.Logger,
	recordRemover RecordRemover,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.records.delete.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var pathParams PathParams

		recordIdStr := chi.URLParam(r, "recordId")
		recordId, err := strconv.ParseInt(recordIdStr, 10, 64)
		if err != nil {
			render.JSON(w, r, response.ErrorMessage("invalid url path record id"))
			return
		}

		pathParams.RecordId = recordId

		if err := validator.New().Struct(pathParams); err != nil {
			validateErr := err.(validator.ValidationErrors)
			log.Info("invalid request", slog.String("err", err.Error()))
			render.JSON(w, r, response.ValidationError(validateErr))
			return
		}

		if err := recordRemover.DeleteRecordForCurrentUser(r.Context(), pathParams.RecordId); err != nil {
			render.JSON(w, r, response.ErrorMessage("unexpected error"))
			return
		}

		render.Status(r, http.StatusNoContent)
		render.JSON(w, r, nil)
	}
}
