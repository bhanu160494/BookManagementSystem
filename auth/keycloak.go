package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"Go-Learning/config"
)

// ErrInvalidCredentials is returned when Keycloak rejects a username/password pair.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrUserExists is returned when a signup conflicts with an existing account.
var ErrUserExists = errors.New("user already exists")

// TokenResponse mirrors the OIDC token endpoint payload returned to API clients.
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token,omitempty"`
	RefreshExpiresIn int    `json:"refresh_expires_in,omitempty"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope,omitempty"`
}

// KeycloakClient wraps the subset of the Keycloak REST API used by this service.
type KeycloakClient struct {
	cfg        config.KeycloakConfig
	httpClient *http.Client

	adminMu      sync.Mutex
	adminToken   string
	adminExpires time.Time
}

// NewKeycloakClient creates a client for the supplied realm configuration.
func NewKeycloakClient(cfg config.KeycloakConfig) *KeycloakClient {
	return &KeycloakClient{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// Config exposes the realm configuration used by this client.
func (k *KeycloakClient) Config() config.KeycloakConfig {
	return k.cfg
}

// Login exchanges user credentials for an access/refresh token pair.
func (k *KeycloakClient) Login(ctx context.Context, username, password string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", k.cfg.ClientID)
	form.Set("username", username)
	form.Set("password", password)
	form.Set("scope", "openid profile email")
	if k.cfg.ClientSecret != "" {
		form.Set("client_secret", k.cfg.ClientSecret)
	}
	return k.requestToken(ctx, k.cfg.TokenURL(), form)
}

// Refresh exchanges a refresh token for a new access token.
func (k *KeycloakClient) Refresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", k.cfg.ClientID)
	form.Set("refresh_token", refreshToken)
	if k.cfg.ClientSecret != "" {
		form.Set("client_secret", k.cfg.ClientSecret)
	}
	return k.requestToken(ctx, k.cfg.TokenURL(), form)
}

// Logout revokes the supplied refresh token and ends the Keycloak session.
func (k *KeycloakClient) Logout(ctx context.Context, refreshToken string) error {
	form := url.Values{}
	form.Set("client_id", k.cfg.ClientID)
	form.Set("refresh_token", refreshToken)
	if k.cfg.ClientSecret != "" {
		form.Set("client_secret", k.cfg.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, k.cfg.LogoutURL(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("keycloak logout failed with status %d", resp.StatusCode)
	}
	return nil
}

// NewUser describes the account to create during signup.
type NewUser struct {
	Username  string
	Email     string
	FirstName string
	LastName  string
	Password  string
}

// CreateUser registers a new user in the realm and assigns the default role.
func (k *KeycloakClient) CreateUser(ctx context.Context, user NewUser) (string, error) {
	adminToken, err := k.adminAccessToken(ctx)
	if err != nil {
		return "", err
	}

	payload := map[string]any{
		"username":      user.Username,
		"email":         user.Email,
		"firstName":     user.FirstName,
		"lastName":      user.LastName,
		"enabled":       true,
		"emailVerified": false,
		"credentials": []map[string]any{
			{"type": "password", "value": user.Password, "temporary": false},
		},
	}

	resp, err := k.doAdminJSON(ctx, http.MethodPost, k.cfg.AdminAPIURL("/users"), adminToken, payload)
	if err != nil {
		return "", err
	}
	defer drainAndClose(resp.Body)

	switch {
	case resp.StatusCode == http.StatusConflict:
		return "", ErrUserExists
	case resp.StatusCode >= 400:
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return "", fmt.Errorf("keycloak user creation failed (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	userID := userIDFromLocation(resp.Header.Get("Location"))
	if userID == "" {
		userID, err = k.findUserID(ctx, adminToken, user.Username)
		if err != nil {
			return "", err
		}
	}

	if k.cfg.DefaultUserRole != "" {
		if err := k.assignRealmRole(ctx, adminToken, userID, k.cfg.DefaultUserRole); err != nil {
			return userID, err
		}
	}
	return userID, nil
}

func (k *KeycloakClient) findUserID(ctx context.Context, adminToken, username string) (string, error) {
	endpoint := k.cfg.AdminAPIURL("/users?exact=true&username=" + url.QueryEscape(username))
	resp, err := k.doAdminJSON(ctx, http.MethodGet, endpoint, adminToken, nil)
	if err != nil {
		return "", err
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("keycloak user lookup failed with status %d", resp.StatusCode)
	}

	var users []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return "", err
	}
	if len(users) == 0 {
		return "", errors.New("created user could not be found in keycloak")
	}
	return users[0].ID, nil
}

func (k *KeycloakClient) assignRealmRole(ctx context.Context, adminToken, userID, roleName string) error {
	roleResp, err := k.doAdminJSON(ctx, http.MethodGet, k.cfg.AdminAPIURL("/roles/"+url.PathEscape(roleName)), adminToken, nil)
	if err != nil {
		return err
	}
	defer drainAndClose(roleResp.Body)

	if roleResp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("realm role %q does not exist in keycloak", roleName)
	}
	if roleResp.StatusCode >= 400 {
		return fmt.Errorf("keycloak role lookup failed with status %d", roleResp.StatusCode)
	}

	var role struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(roleResp.Body).Decode(&role); err != nil {
		return err
	}

	endpoint := k.cfg.AdminAPIURL("/users/" + url.PathEscape(userID) + "/role-mappings/realm")
	assignResp, err := k.doAdminJSON(ctx, http.MethodPost, endpoint, adminToken, []map[string]string{
		{"id": role.ID, "name": role.Name},
	})
	if err != nil {
		return err
	}
	defer drainAndClose(assignResp.Body)

	if assignResp.StatusCode >= 400 {
		return fmt.Errorf("keycloak role assignment failed with status %d", assignResp.StatusCode)
	}
	return nil
}

func (k *KeycloakClient) adminAccessToken(ctx context.Context) (string, error) {
	k.adminMu.Lock()
	defer k.adminMu.Unlock()

	if k.adminToken != "" && time.Now().Before(k.adminExpires) {
		return k.adminToken, nil
	}

	if k.cfg.AdminPassword == "" {
		return "", errors.New("KEYCLOAK_ADMIN_PASSWORD is not configured")
	}

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("client_id", k.cfg.AdminClientID)
	form.Set("username", k.cfg.AdminUsername)
	form.Set("password", k.cfg.AdminPassword)

	token, err := k.requestToken(ctx, k.cfg.AdminTokenURL(), form)
	if err != nil {
		return "", fmt.Errorf("unable to authenticate keycloak admin: %w", err)
	}

	k.adminToken = token.AccessToken
	// Refresh slightly before the real expiry to avoid races on long requests.
	k.adminExpires = time.Now().Add(time.Duration(maxInt(token.ExpiresIn-30, 10)) * time.Second)
	return k.adminToken, nil
}

func (k *KeycloakClient) requestToken(ctx context.Context, endpoint string, form url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest {
		return nil, ErrInvalidCredentials
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("keycloak token request failed with status %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	return &token, nil
}

func (k *KeycloakClient) doAdminJSON(ctx context.Context, method, endpoint, adminToken string, payload any) (*http.Response, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return k.httpClient.Do(req)
}

func userIDFromLocation(location string) string {
	if location == "" {
		return ""
	}
	parts := strings.Split(strings.TrimRight(location, "/"), "/")
	return parts[len(parts)-1]
}

func drainAndClose(body io.ReadCloser) {
	_, _ = io.Copy(io.Discard, io.LimitReader(body, 4096))
	_ = body.Close()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
