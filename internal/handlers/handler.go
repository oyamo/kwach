package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/oyamo/kwach/internal/application"
	"github.com/oyamo/kwach/internal/domain"
)

const (
	maxRequestBytes = 16 << 10
	roastsPerClient = 5
	roastsPerIP     = 20
	roastRateWindow = 10 * time.Minute
	roastRetryAfter = "600"
)

type Handler struct {
	roaster       application.Roaster
	logger        *slog.Logger
	clientLimiter *limiter
	ipLimiter     *limiter
	allowedOrigin string
}

type profileInput struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Details     string `json:"details"`
	Intensity   string `json:"intensity"`
}

type profileOutput struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Subcategory string `json:"subcategory"`
	Details     string `json:"details"`
	Intensity   string `json:"intensity"`
}

type errorOutput struct {
	Error string `json:"error"`
}

func NewHandler(roaster application.Roaster, logger *slog.Logger) Handler {
	return Handler{
		roaster:       roaster,
		logger:        logger,
		clientLimiter: newLimiter(roastsPerClient, roastRateWindow),
		ipLimiter:     newLimiter(roastsPerIP, roastRateWindow),
	}
}

func (h Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/profiles/{clientID}", h.getProfile)
	mux.HandleFunc("PUT /api/profiles/{clientID}", h.saveProfile)
	mux.HandleFunc("POST /api/profiles/{clientID}/roast", h.limitRoasts(h.roast))
	mux.HandleFunc("DELETE /api/profiles/{clientID}/history", h.clearHistory)
	return h.withCORS(h.logRequests(mux))
}

func (h Handler) clearHistory(w http.ResponseWriter, r *http.Request) {
	if err := h.roaster.ClearHistory(r.Context(), r.PathValue("clientID")); err != nil {
		h.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	profile, err := h.roaster.GetProfile(r.Context(), r.PathValue("clientID"))
	if err != nil {
		h.fail(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, profileOutput{
		Name:        profile.Name,
		Category:    string(profile.Category),
		Subcategory: profile.Subcategory,
		Details:     profile.Details,
		Intensity:   string(profile.Intensity),
	})
}

func (h Handler) saveProfile(w http.ResponseWriter, r *http.Request) {
	var input profileInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	profile := domain.Profile{
		ClientID:    r.PathValue("clientID"),
		Name:        input.Name,
		Category:    domain.Category(input.Category),
		Subcategory: input.Subcategory,
		Details:     input.Details,
		Intensity:   domain.Intensity(input.Intensity),
	}
	if err := h.roaster.SaveProfile(r.Context(), profile); err != nil {
		h.fail(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) roast(w http.ResponseWriter, r *http.Request) {
	audio, err := h.roaster.Roast(r.Context(), r.PathValue("clientID"))
	if err != nil {
		h.fail(w, r, err)
		return
	}
	w.Header().Set("Content-Type", audio.ContentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(audio.Data)))
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(audio.Data)
}

func (h Handler) limitRoasts(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.clientLimiter.allow(r.PathValue("clientID")) || !h.ipLimiter.allow(clientIP(r)) {
			w.Header().Set("Retry-After", roastRetryAfter)
			writeError(w, http.StatusTooManyRequests, "too many roasts, try again later")
			return
		}
		next(w, r)
	}
}

func (h Handler) fail(w http.ResponseWriter, r *http.Request, err error) {
	status := statusFor(err)
	if status >= http.StatusInternalServerError {
		h.logger.Error("request failed", "route", r.Pattern, "error", err)
		writeError(w, status, "internal error")
		return
	}
	writeError(w, status, err.Error())
}

func (h Handler) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(recorder, r)
		h.logger.Info("request handled",
			"method", r.Method,
			"route", r.Pattern,
			"status", recorder.status,
			"duration", time.Since(started),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(status int) {
	s.status = status
	s.ResponseWriter.WriteHeader(status)
}

func statusFor(err error) int {
	switch {
	case errors.Is(err, domain.ErrProfileNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrUnsafeScript):
		return http.StatusUnprocessableEntity
	case errors.Is(err, domain.ErrInvalidClientID), errors.Is(err, domain.ErrInvalidProfile),
		errors.Is(err, domain.ErrEmptyPrompt), errors.Is(err, domain.ErrPromptTooLong):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorOutput{Error: message})
}
