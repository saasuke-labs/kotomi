package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/oauth2"
)

// Auth0Config holds Auth0 configuration
type Auth0Config struct {
	Domain       string
	ClientID     string
	ClientSecret string
	CallbackURL  string
	OAuth2Config *oauth2.Config
}

// UserInfo represents user information from Auth0
type UserInfo struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Picture       string `json:"picture"`
}

// NewAuth0Config creates a new Auth0 configuration from environment variables
func NewAuth0Config() (*Auth0Config, error) {
	domain := os.Getenv("AUTH0_DOMAIN")
	if domain == "" {
		return nil, fmt.Errorf("AUTH0_DOMAIN environment variable is required")
	}

	clientID := os.Getenv("AUTH0_CLIENT_ID")
	if clientID == "" {
		return nil, fmt.Errorf("AUTH0_CLIENT_ID environment variable is required")
	}

	clientSecret := os.Getenv("AUTH0_CLIENT_SECRET")
	if clientSecret == "" {
		return nil, fmt.Errorf("AUTH0_CLIENT_SECRET environment variable is required")
	}

	callbackURL := os.Getenv("AUTH0_CALLBACK_URL")
	if callbackURL == "" {
		callbackURL = "http://localhost:8080/callback"
	}

	conf := &Auth0Config{
		Domain:       domain,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CallbackURL:  callbackURL,
	}

	conf.OAuth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  callbackURL,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://%s/authorize", domain),
			TokenURL: fmt.Sprintf("https://%s/oauth/token", domain),
		},
	}

	return conf, nil
}

// GetLoginURL generates the Auth0 login URL with state parameter
func (c *Auth0Config) GetLoginURL(state string) string {
	return c.OAuth2Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

// ExchangeCode exchanges an authorization code for tokens
func (c *Auth0Config) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := c.OAuth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	return token, nil
}

// GetUserInfo fetches user information from Auth0
func (c *Auth0Config) GetUserInfo(ctx context.Context, token *oauth2.Token) (*UserInfo, error) {
	client := c.OAuth2Config.Client(ctx, token)
	userInfoURL := fmt.Sprintf("https://%s/userinfo", c.Domain)

	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status: %d", resp.StatusCode)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// GetLogoutURL generates the Auth0 logout URL
func (c *Auth0Config) GetLogoutURL(returnTo string) string {
	logoutURL := fmt.Sprintf("https://%s/v2/logout", c.Domain)
	params := url.Values{}
	params.Add("client_id", c.ClientID)
	params.Add("returnTo", returnTo)
	return fmt.Sprintf("%s?%s", logoutURL, params.Encode())
}

// GenerateRandomState generates a random state parameter for OAuth flow
func GenerateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
