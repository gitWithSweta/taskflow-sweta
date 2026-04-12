package handler

import (
	"errors"
	"log/slog"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"

	"taskflow/internal/errs"
	"taskflow/internal/httpx"
)

func writeErr(w http.ResponseWriter, r *http.Request, log *slog.Logger, err error) {
	if err == nil {
		return
	}
	var v *errs.ValidationError
	if errors.As(err, &v) {
		httpx.WriteValidation(w, v.Fields)
		return
	}
	switch {
	case errors.Is(err, errs.ErrNotFound):
		httpx.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "not found"})
	case errors.Is(err, errs.ErrForbidden):
		httpx.WriteJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
	case errors.Is(err, errs.ErrInvalidCredentials), errors.Is(err, errs.ErrUnauthorized):
		httpx.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	default:
		if log != nil {
			log.ErrorContext(r.Context(), "internal_error",
				slog.String("request_id", chimw.GetReqID(r.Context())),
				slog.String("err", err.Error()),
			)
		}
		httpx.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal error"})
	}
}
