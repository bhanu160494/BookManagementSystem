package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"Go-Learning/config"

	"github.com/golang-jwt/jwt/v5"
)

// minJWKSRefreshInterval throttles key-set refreshes so an attacker cannot force
// unbounded outbound requests by sending tokens with unknown `kid` values.
const minJWKSRefreshInterval = 30 * time.Second

// Claims holds the token attributes the API cares about.
type Claims struct {
	Subject  string
	Username string
	Email    string
	Roles    []string
}

// HasRole reports whether the token carries the given realm/client role.
type roleAccess struct {
	Roles []string `json:"roles"`
}

type keycloakClaims struct {
	jwt.RegisteredClaims
	PreferredUsername string                `json:"preferred_username"`
	Email             string                `json:"email"`
	RealmAccess       roleAccess            `json:"realm_access"`
	ResourceAccess    map[string]roleAccess `json:"resource_access"`
	AuthorizedParty   string                `json:"azp"`
}

// TokenVerifier validates Keycloak access tokens offline using the realm JWKS.
type TokenVerifier struct {
	cfg        config.KeycloakConfig
	httpClient *http.Client

	mu          sync.RWMutex
	keys        map[string]*rsa.PublicKey
	lastRefresh time.Time
}

// NewTokenVerifier builds a verifier for the configured realm.
func NewTokenVerifier(cfg config.KeycloakConfig) *TokenVerifier {
	return &TokenVerifier{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		keys:       map[string]*rsa.PublicKey{},
	}
}

// Verify parses and validates a raw bearer token and returns its claims.
func (v *TokenVerifier) Verify(ctx context.Context, rawToken string) (*Claims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
		jwt.WithIssuer(v.cfg.Issuer()),
		jwt.WithExpirationRequired(),
	)

	claims := &keycloakClaims{}
	_, err := parser.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("token is missing a key id")
		}
		return v.publicKey(ctx, kid)
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if v.cfg.ExpectedAudience != "" && !claims.VerifyAudienceOrParty(v.cfg.ExpectedAudience) {
		return nil, errors.New("invalid token: unexpected audience")
	}

	roles := append([]string{}, claims.RealmAccess.Roles...)
	if clientRoles, ok := claims.ResourceAccess[v.cfg.ClientID]; ok {
		roles = append(roles, clientRoles.Roles...)
	}

	return &Claims{
		Subject:  claims.Subject,
		Username: claims.PreferredUsername,
		Email:    claims.Email,
		Roles:    roles,
	}, nil
}

// VerifyAudienceOrParty accepts the token when either `aud` or `azp` matches.
func (c *keycloakClaims) VerifyAudienceOrParty(expected string) bool {
	if c.AuthorizedParty == expected {
		return true
	}
	for _, aud := range c.Audience {
		if aud == expected {
			return true
		}
	}
	return false
}

func (v *TokenVerifier) publicKey(ctx context.Context, kid string) (*rsa.PublicKey, error) {
	v.mu.RLock()
	key, ok := v.keys[kid]
	stale := time.Since(v.lastRefresh) > minJWKSRefreshInterval
	v.mu.RUnlock()

	if ok {
		return key, nil
	}
	if !stale {
		return nil, fmt.Errorf("unknown signing key %q", kid)
	}

	if err := v.refreshKeys(ctx); err != nil {
		return nil, err
	}

	v.mu.RLock()
	defer v.mu.RUnlock()
	if key, ok := v.keys[kid]; ok {
		return key, nil
	}
	return nil, fmt.Errorf("unknown signing key %q", kid)
}

func (v *TokenVerifier) refreshKeys(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.cfg.JWKSURL(), nil)
	if err != nil {
		return err
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to fetch keycloak jwks: %w", err)
	}
	defer drainAndClose(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("keycloak jwks request failed with status %d", resp.StatusCode)
	}

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			Use string `json:"use"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	keys := map[string]*rsa.PublicKey{}
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" || (k.Use != "" && k.Use != "sig") {
			continue
		}
		pub, err := rsaPublicKey(k.N, k.E)
		if err != nil {
			continue
		}
		keys[k.Kid] = pub
	}
	if len(keys) == 0 {
		return errors.New("keycloak jwks contained no usable RSA signing keys")
	}

	v.mu.Lock()
	v.keys = keys
	v.lastRefresh = time.Now()
	v.mu.Unlock()
	return nil
}

func rsaPublicKey(modulus, exponent string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(modulus)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(exponent)
	if err != nil {
		return nil, err
	}
	if len(eBytes) == 0 || len(eBytes) > 8 {
		return nil, errors.New("unsupported rsa exponent size")
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: int(new(big.Int).SetBytes(eBytes).Int64()),
	}, nil
}
