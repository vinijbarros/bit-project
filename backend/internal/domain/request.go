package domain

import "time"

const (
	CategoryTI             = "ti"
	CategoryRH             = "rh"
	CategoryPurchases      = "compras"
	CategoryFinance        = "financeiro"
	CategoryInfrastructure = "infraestrutura"

	StatusOpen       = "aberto"
	StatusInProgress = "em_atendimento"
	StatusCompleted  = "concluido"
)

var categoryLabels = map[string]string{
	CategoryTI:             "TI",
	CategoryRH:             "RH",
	CategoryPurchases:      "Compras",
	CategoryFinance:        "Financeiro",
	CategoryInfrastructure: "Infraestrutura",
}

var statusLabels = map[string]string{
	StatusOpen:       "Aberto",
	StatusInProgress: "Em Atendimento",
	StatusCompleted:  "Concluído",
}

type Request struct {
	ID          int64
	Title       string
	Description string
	Category    string
	Status      string
	Requester   User
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type LabeledValue struct {
	Value string
	Label string
}

func Categories() []LabeledValue {
	return []LabeledValue{
		{Value: CategoryTI, Label: categoryLabels[CategoryTI]},
		{Value: CategoryRH, Label: categoryLabels[CategoryRH]},
		{Value: CategoryPurchases, Label: categoryLabels[CategoryPurchases]},
		{Value: CategoryFinance, Label: categoryLabels[CategoryFinance]},
		{Value: CategoryInfrastructure, Label: categoryLabels[CategoryInfrastructure]},
	}
}

func Statuses() []LabeledValue {
	return []LabeledValue{
		{Value: StatusOpen, Label: statusLabels[StatusOpen]},
		{Value: StatusInProgress, Label: statusLabels[StatusInProgress]},
		{Value: StatusCompleted, Label: statusLabels[StatusCompleted]},
	}
}

func CategoryLabel(value string) (string, bool) {
	label, ok := categoryLabels[value]
	return label, ok
}

func StatusLabel(value string) (string, bool) {
	label, ok := statusLabels[value]
	return label, ok
}
