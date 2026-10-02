package httpapi

import (
	"net/http"
)

func requireTrustedOrigin(trustedOrigins map[string]struct{}, next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if !changesState(request.Method) {
			next.ServeHTTP(response, request)
			return
		}
		if _, trusted := trustedOrigins[request.Header.Get("Origin")]; !trusted {
			writeError(response, request, http.StatusForbidden, "origin_not_allowed", "Origem da requisição não permitida.", nil)
			return
		}
		next.ServeHTTP(response, request)
	})
}

func trustedOriginSet(origins []string) map[string]struct{} {
	set := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		set[origin] = struct{}{}
	}
	return set
}

func changesState(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
