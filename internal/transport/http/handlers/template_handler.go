package handlers

import (
	"net/http"
	"time"

	templateusecase "example.com/taskservice/internal/usecase/template"
)

type TemplateHandler struct {
	usecase templateusecase.Usecase
}

func NewTemplateHandler(usecase templateusecase.Usecase) *TemplateHandler {
	return &TemplateHandler{usecase: usecase}
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createTemplateDTO

	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		parsed, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		endDate = &parsed
	}

	var dates []time.Time
	for _, d := range req.SpecificDates {
		parsed, err := time.Parse(time.RFC3339, d)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		dates = append(dates, parsed)
	}

	err = h.usecase.Create(r.Context(), templateusecase.CreateInput{
		Title:          req.Title,
		Description:    req.Description,
		RecurrenceType: req.RecurrenceType,
		Interval:       req.Interval,
		DaysOfMonth:    req.DaysOfMonth,
		SpecificDates:  dates,
		Parity:         req.Parity,
		StartDate:      startDate,
		EndDate:        endDate,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
