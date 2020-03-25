package auth

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"fmt"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

type User struct {
	// TODO: Do we actually need email? Not unique according to spec.
	ID        string `json:"sub"`
	Namespace string
}

const (
	ClientID     = "b87He1pQEDohVzOAYAfLIUfixO5zu6Ln"
	AuthHost     = "https://blimp-testing.auth0.com"
	AuthURL      = AuthHost + "/authorize"
	TokenURL     = AuthHost + "/oauth/token"
	RedirectHost = "localhost:8085"
	RedirectPath = "/oauth/redirect"
)

var (
	// The base64 encoded certificate for the cluster manager. This is set at build time.
	ClusterManagerCertBase64 string

	// The PEM-encoded certificate for the cluster manager.
	ClusterManagerCert = mustDecodeBase64(ClusterManagerCertBase64)
)

var Endpoint = oauth2.Endpoint{
	AuthURL:   AuthHost + "/authorize",
	TokenURL:  AuthHost + "/oauth/token",
	AuthStyle: oauth2.AuthStyleInParams,
}

var verifier = oidc.NewVerifier(
	"https://blimp-testing.auth0.com/",
	// TODO: Fetching over the network.. Any issues if no network connectivity?
	oidc.NewRemoteKeySet(context.Background(), "https://blimp-testing.auth0.com/.well-known/jwks.json"),
	&oidc.Config{ClientID: ClientID})

func ParseIDToken(token string) (User, error) {
	idToken, err := verifier.Verify(context.Background(), token)
	if err != nil {
		return User{}, fmt.Errorf("verify: %w", err)
	}

	var user User
	if err := idToken.Claims(&user); err != nil {
		return User{}, fmt.Errorf("parse claims: %w", err)
	}

	user.Namespace = dnsCompliantHash(user.ID)
	return user, nil
}

func mustDecodeBase64(encoded string) string {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err)
	}
	return string(decoded)
}

// dnsCompliantHash hashes the given string and encodes it into base16.
func dnsCompliantHash(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}
