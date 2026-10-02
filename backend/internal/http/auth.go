package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"portal-solicitacoes/internal/domain"
	"portal-solicitacoes/internal/service"
)

type authService interface {
	Login(context.Context, service.LoginInput, string) (service.LoginResult, error)
	Authenticate(context.Context, string) (domain.User, error)
	Logout(context.Context, string) error
}

type sessionCookieConfig struct {
	Name     string
	Path     string
	Secure   bool
	Duration time.Duration
}

func SessionCookieConfig(name string, secure bool, duration time.Duration) sessionCookieConfig {
	return sessionCookieConfig{Name: name, Path: "/", Secure: secure, Duration: duration}
}

type authHandler struct {
	service authService
	cookie  sessionCookieConfig
	limiter *loginRateLimiter
	logger  *slog.Logger
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type userResponse struct {
	Data userData `json:"data"`
}

type userData struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

func (handler *authHandler) login(response http.ResponseWriter, request *http.Request) {
	var input loginRequest
	if err := readJSON(response, request, &input); err != nil {
		writeRequestError(response, request, err)
		return
	}

	limitKey := loginLimitKey(request, input.Username)
	if allowed, retryAfter := handler.limiter.allow(limitKey); !allowed {
		response.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
		writeError(response, request, http.StatusTooManyRequests, "login_rate_limited", "Muitas tentativas de login. Tente novamente mais tarde.", nil)
		return
	}

	previousToken := cookieToken(request, handler.cookie.Name)
	result, err := handler.service.Login(request.Context(), service.LoginInput{
		Username: input.Username,
		Password: input.Password,
	}, previousToken)
	if err != nil {
		var validationErr *service.ValidationError
		switch {
		case errors.As(err, &validationErr):
			handler.limiter.failure(limitKey)
			writeError(response, request, http.StatusBadRequest, "validation_failed", "Corrija os campos informados.", fieldErrors(validationErr.Fields))
		case errors.Is(err, service.ErrInvalidCredentials):
			handler.limiter.failure(limitKey)
			writeError(response, request, http.StatusUnauthorized, "invalid_credentials", "Usuário ou senha inválidos.", nil)
		default:
			writeInternalError(handler.logger, response, request, err)
		}
		return
	}

	handler.limiter.reset(limitKey)
	handler.setSessionCookie(response, result.Token, result.ExpiresAt)
	writeJSON(response, http.StatusOK, userResponse{Data: presentUser(result.User)})
}

func (handler *authHandler) me(response http.ResponseWriter, request *http.Request) {
	user, ok := authenticatedUser(request.Context())
	if !ok {
		writeError(response, request, http.StatusInternalServerError, "internal_error", "Erro interno inesperado.", nil)
		return
	}
	writeJSON(response, http.StatusOK, userResponse{Data: presentUser(user)})
}

func (handler *authHandler) logout(response http.ResponseWriter, request *http.Request) {
	token := cookieToken(request, handler.cookie.Name)
	if err := handler.service.Logout(request.Context(), token); err != nil {
		writeInternalError(handler.logger, response, request, err)
		return
	}
	handler.clearSessionCookie(response)
	response.Header().Set("Cache-Control", "no-store")
	response.WriteHeader(http.StatusNoContent)
}

func (handler *authHandler) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		token := cookieToken(request, handler.cookie.Name)
		user, err := handler.service.Authenticate(request.Context(), token)
		if errors.Is(err, service.ErrUnauthenticated) {
			writeError(response, request, http.StatusUnauthorized, "authentication_required", "Autenticação necessária.", nil)
			return
		}
		if err != nil {
			writeInternalError(handler.logger, response, request, err)
			return
		}
		ctx := context.WithValue(request.Context(), authenticatedUserContextKey, user)
		next.ServeHTTP(response, request.WithContext(ctx))
	})
}

func (handler *authHandler) setSessionCookie(response http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(response, &http.Cookie{
		Name:     handler.cookie.Name,
		Value:    token,
		Path:     handler.cookie.Path,
		Expires:  expiresAt,
		MaxAge:   int(handler.cookie.Duration.Seconds()),
		HttpOnly: true,
		Secure:   handler.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (handler *authHandler) clearSessionCookie(response http.ResponseWriter) {
	http.SetCookie(response, &http.Cookie{
		Name:     handler.cookie.Name,
		Value:    "",
		Path:     handler.cookie.Path,
		Expires:  time.Unix(1, 0).UTC(),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   handler.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func cookieToken(request *http.Request, name string) string {
	cookie, err := request.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func presentUser(user domain.User) userData {
	return userData{ID: user.ID, Username: user.Username, DisplayName: user.DisplayName}
}

const authenticatedUserContextKey contextKey = "authenticated_user"

func authenticatedUser(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(authenticatedUserContextKey).(domain.User)
	return user, ok
}
