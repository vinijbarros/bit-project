package service

import (
	"context"
	"errors"
	"testing"

	"portal-solicitacoes/internal/domain"
)

type stubDashboardRepository struct {
	counts domain.DashboardCounts
	err    error
}

func (stub stubDashboardRepository) Counts(context.Context) (domain.DashboardCounts, error) {
	return stub.counts, stub.err
}

func TestDashboardReturnsConsistentCounts(t *testing.T) {
	want := domain.DashboardCounts{Total: 7, Open: 2, InProgress: 3, Completed: 2}
	got, err := NewDashboard(stubDashboardRepository{counts: want}).Get(context.Background())
	if err != nil || got != want {
		t.Fatalf("Get() = %+v, %v; want %+v, nil", got, err, want)
	}
}

func TestDashboardDoesNotTurnRepositoryFailureIntoZeros(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	got, err := NewDashboard(stubDashboardRepository{err: databaseErr}).Get(context.Background())
	if !errors.Is(err, databaseErr) || got != (domain.DashboardCounts{}) {
		t.Fatalf("Get() = %+v, %v; want zero value and repository error", got, err)
	}
}

func TestDashboardRejectsInconsistentCounts(t *testing.T) {
	_, err := NewDashboard(stubDashboardRepository{counts: domain.DashboardCounts{
		Total: 4, Open: 1, InProgress: 1, Completed: 1,
	}}).Get(context.Background())
	if !errors.Is(err, ErrDashboardInconsistent) {
		t.Fatalf("Get() error = %v; want ErrDashboardInconsistent", err)
	}
}
