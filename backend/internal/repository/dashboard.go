package repository

import (
	"context"
	"database/sql"

	"portal-solicitacoes/internal/domain"
)

type Dashboard struct {
	db *sql.DB
}

func NewDashboard(db *sql.DB) *Dashboard {
	return &Dashboard{db: db}
}

func (repository *Dashboard) Counts(ctx context.Context) (domain.DashboardCounts, error) {
	const query = `
		SELECT
			COUNT(*)::bigint,
			COUNT(*) FILTER (WHERE status = 'aberto')::bigint,
			COUNT(*) FILTER (WHERE status = 'em_atendimento')::bigint,
			COUNT(*) FILTER (WHERE status = 'concluido')::bigint
		FROM requests`

	var counts domain.DashboardCounts
	err := repository.db.QueryRowContext(ctx, query).Scan(
		&counts.Total,
		&counts.Open,
		&counts.InProgress,
		&counts.Completed,
	)
	return counts, err
}
