package middleware

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	RefreshTokenHeader = "X-Refresh-Token"
	KeycloakClaimsKey  = "keycloak_claims"
)

type KeycloakConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client
}

type KeycloakClaims struct {
	Subject           string `json:"sub"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	AuthorizedParty   string `json:"azp"`
	ResourceAccess    map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
}

type keycloakTokenResponse struct {
	AccessToken string `json:"access_token"`
}

func KeycloakRefreshTokenMiddleware(cfg KeycloakConfig) fiber.Handler {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}

	return func(c *fiber.Ctx) error {
		refreshToken := c.Get(RefreshTokenHeader)
		if refreshToken == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "refresh token отсутствует")
		}

		token, err := refreshAccessToken(c.UserContext(), httpClient, cfg, refreshToken)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "refresh token недействителен")
		}

		claims, err := parseAccessTokenClaims(token.AccessToken)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "access token недействителен")
		}

		c.Locals(KeycloakClaimsKey, claims)

		return c.Next()
	}
}

func RequireClientRole(clientID string, role string) fiber.Handler {
	return func(fiberCtx *fiber.Ctx) error {
		claims, err := ClaimsFromContext(fiberCtx)
		if err != nil {
			return err
		}

		if !claims.HasClientRole(clientID, role) {
			return fiber.NewError(fiber.StatusForbidden, "доступ запрещен")
		}

		return fiberCtx.Next()
	}
}

func ClaimsFromContext(fiberCtx *fiber.Ctx) (*KeycloakClaims, error) {
	value := fiberCtx.Locals(KeycloakClaimsKey)

	claims, ok := value.(*KeycloakClaims)
	if !ok {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "claims токена отсутствуют")
	}

	return claims, nil
}

func (claims KeycloakClaims) HasClientRole(clientID string, role string) bool {
	access, ok := claims.ResourceAccess[clientID]
	if !ok {
		return false
	}

	for _, currentRole := range access.Roles {
		if currentRole == role {
			return true
		}
	}

	return false
}

func refreshAccessToken(
	ctx context.Context,
	httpClient *http.Client,
	cfg KeycloakConfig,
	refreshToken string,
) (*keycloakTokenResponse, error) {
	if cfg.Issuer == "" {
		return nil, errors.New("keycloak issuer is required")
	}
	if cfg.ClientID == "" {
		return nil, errors.New("keycloak client id is required")
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", cfg.ClientID)
	form.Set("refresh_token", refreshToken)

	if cfg.ClientSecret != "" {
		form.Set("client_secret", cfg.ClientSecret)
	}

	endpoint := strings.TrimRight(cfg.Issuer, "/") + "/protocol/openid-connect/token"

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoint,
		bytes.NewBufferString(form.Encode()),
	)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, response.Body)
		return nil, fmt.Errorf("keycloak token endpoint returned status %d", response.StatusCode)
	}

	var token keycloakTokenResponse
	if err := json.NewDecoder(response.Body).Decode(&token); err != nil {
		return nil, err
	}

	if token.AccessToken == "" {
		return nil, errors.New("keycloak response does not contain access_token")
	}

	return &token, nil
}

func parseAccessTokenClaims(accessToken string) (*KeycloakClaims, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil, errors.New("jwt must contain three parts")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims KeycloakClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	if claims.Subject == "" {
		return nil, errors.New("jwt does not contain sub claim")
	}

	return &claims, nil
}
