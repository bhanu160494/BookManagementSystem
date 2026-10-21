package config

import (
	"os"
	"strings"
)

func GetEnvValue(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// KeycloakConfig holds every setting needed to talk to a Keycloak realm.
type KeycloakConfig struct {
	BaseURL          string
	InternalBaseURL  string
	Realm            string
	ClientID         string
	ClientSecret     string
	AdminRealm       string
	AdminClientID    string
	AdminUsername    string
	AdminPassword    string
	DefaultUserRole  string
	ExpectedAudience string
}

// LoadKeycloakConfig reads the Keycloak settings from the environment.
func LoadKeycloakConfig() KeycloakConfig {
	baseURL := strings.TrimRight(GetEnvValue("KEYCLOAK_BASE_URL", "http://localhost:8081"), "/")
	return KeycloakConfig{
		BaseURL: baseURL,
		// Set this when the service reaches Keycloak over a private network
		// address that differs from the public issuer URL (e.g. in Docker).
		InternalBaseURL:  strings.TrimRight(GetEnvValue("KEYCLOAK_INTERNAL_BASE_URL", baseURL), "/"),
		Realm:            GetEnvValue("KEYCLOAK_REALM", "bookstore"),
		ClientID:         GetEnvValue("KEYCLOAK_CLIENT_ID", "bookstore-api"),
		ClientSecret:     os.Getenv("KEYCLOAK_CLIENT_SECRET"),
		AdminRealm:       GetEnvValue("KEYCLOAK_ADMIN_REALM", "master"),
		AdminClientID:    GetEnvValue("KEYCLOAK_ADMIN_CLIENT_ID", "admin-cli"),
		AdminUsername:    GetEnvValue("KEYCLOAK_ADMIN_USERNAME", "admin"),
		AdminPassword:    os.Getenv("KEYCLOAK_ADMIN_PASSWORD"),
		DefaultUserRole:  GetEnvValue("KEYCLOAK_DEFAULT_USER_ROLE", "book-user"),
		ExpectedAudience: os.Getenv("KEYCLOAK_EXPECTED_AUDIENCE"),
	}
}

// Issuer is the expected `iss` claim value for tokens minted by the realm.
func (c KeycloakConfig) Issuer() string {
	return c.BaseURL + "/realms/" + c.Realm
}

func (c KeycloakConfig) internalRealmURL() string {
	return c.InternalBaseURL + "/realms/" + c.Realm
}

// TokenURL is the OIDC token endpoint of the application realm.
func (c KeycloakConfig) TokenURL() string {
	return c.internalRealmURL() + "/protocol/openid-connect/token"
}

// LogoutURL is the OIDC logout endpoint of the application realm.
func (c KeycloakConfig) LogoutURL() string {
	return c.internalRealmURL() + "/protocol/openid-connect/logout"
}

// JWKSURL is the endpoint exposing the realm's public signing keys.
func (c KeycloakConfig) JWKSURL() string {
	return c.internalRealmURL() + "/protocol/openid-connect/certs"
}

// AdminTokenURL is the token endpoint used to obtain an admin access token.
func (c KeycloakConfig) AdminTokenURL() string {
	return c.InternalBaseURL + "/realms/" + c.AdminRealm + "/protocol/openid-connect/token"
}

// AdminAPIURL builds an absolute Admin REST API URL for the application realm.
func (c KeycloakConfig) AdminAPIURL(path string) string {
	return c.InternalBaseURL + "/admin/realms/" + c.Realm + path
}
