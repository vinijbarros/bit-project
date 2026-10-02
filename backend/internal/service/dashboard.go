package service

import (
	"context"
	"errors"

	"portal-solicitacoes/internal/domain"
)

var ErrDashboardInconsistent = errors.New("dashboard counts are inconsistent")

type DashboardRepository interface {
	Counts(context.Context) (domain.DashboardCounts, error)
}

type Dashboard struct {
	repository DashboardRepository
}

func NewDashboard(repository DashboardRepository) *Dashboard {
	return &Dashboard{repository: repository}
}

func (service *Dashboard) Get(ctx context.Context) (domain.DashboardCounts, error) {
	counts, err := service.repository.Counts(ctx)
	if err != nil {
		return domain.DashboardCounts{}, err
	}
	if counts.Total < 0 || counts.Open < 0 || counts.InProgress < 0 || counts.Completed < 0 ||
		counts.Total != counts.Open+counts.InProgress+counts.Completed {
		return domain.DashboardCounts{}, ErrDashboardInconsistent
	}
	return counts, nil
}
