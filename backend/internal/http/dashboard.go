package httpapi

import (
	"context"
	"log/slog"
	"net/http"

	"portal-solicitacoes/internal/domain"
)

type dashboardService interface {
	Get(context.Context) (domain.DashboardCounts, error)
}

type dashboardHandler struct {
	service dashboardService
	logger  *slog.Logger
}

type dashboardResponse struct {
	Data dashboardData `json:"data"`
}

type dashboardData struct {
	Total      int64 `json:"total"`
	Open       int64 `json:"abertas"`
	InProgress int64 `json:"em_atendimento"`
	Completed  int64 `json:"concluidas"`
}

func (handler *dashboardHandler) get(response http.ResponseWriter, request *http.Request) {
	counts, err := handler.service.Get(request.Context())
	if err != nil {
		writeInternalError(handler.logger, response, request, err)
		return
	}

	writeJSON(response, http.StatusOK, dashboardResponse{Data: dashboardData{
		Total: counts.Total, Open: counts.Open, InProgress: counts.InProgress, Completed: counts.Completed,
	}})
}
