package httpapi

import (
	"net/http"

	"portal-solicitacoes/internal/domain"
)

type metadataResponse struct {
	Data metadataData `json:"data"`
}

type metadataData struct {
	Categories []labeledValue `json:"categories"`
	Statuses   []labeledValue `json:"statuses"`
}

func metadata(response http.ResponseWriter, _ *http.Request) {
	writeJSON(response, http.StatusOK, metadataResponse{Data: metadataData{
		Categories: presentLabeledValues(domain.Categories()),
		Statuses:   presentLabeledValues(domain.Statuses()),
	}})
}

func presentLabeledValues(values []domain.LabeledValue) []labeledValue {
	result := make([]labeledValue, len(values))
	for index, value := range values {
		result[index] = labeledValue{Value: value.Value, Label: value.Label}
	}
	return result
}
