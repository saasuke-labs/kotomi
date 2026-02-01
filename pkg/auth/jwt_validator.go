package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/saasuke-labs/kotomi/pkg/models"
)

// JWTValidator handles validation of JWT tokens based on site configuration
type JWTValidator struct {
	config *models.SiteAuthConfig
}

// NewJWTValidator creates a new JWT validator with the given site auth config
func NewJWTValidator(config *models.SiteAuthConfig) *JWTValidator {
	return &JWTValidator{config: config}
}

// ValidateToken validates a JWT token and returns the user information
func (v *JWTValidator) ValidateToken(tokenString string) (*models.KotomiUser, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token is empty")
	}

	// Parse token with appropriate validation method
	var token *jwt.Token
	var err error

	switch v.config.JWTValidationType {
	case "hmac":
		token, err = v.validateHMAC(tokenString)
	case "rsa":
		token, err = v.validateRSA(tokenString)
	case "ecdsa":
		token, err = v.validateECDSA(tokenString)
	case "jwks":
		token, err = v.validateJWKS(tokenString)
	default:
		return nil, fmt.Errorf("unsupported JWT validation type: %s", v.config.JWTValidationType)
	}

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate standard claims
	if err := v.validateStandardClaims(claims); err != nil {
		return nil, fmt.Errorf("invalid standard claims: %w", err)
	}

	// Extract user information
	user, err := v.extractUserFromClaims(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user from claims: %w", err)
	}

	return user, nil
}

// validateHMAC validates token using HMAC symmetric key
func (v *JWTValidator) validateHMAC(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.config.JWTSecret), nil
	})
}

// validateRSA validates token using RSA public key
func (v *JWTValidator) validateRSA(tokenString string) (*jwt.Token, error) {
	// Parse public key
	block, _ := pem.Decode([]byte(v.config.JWTPublicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	rsaKey, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return rsaKey, nil
	})
}

// validateECDSA validates token using ECDSA public key
func (v *JWTValidator) validateECDSA(tokenString string) (*jwt.Token, error) {
	// Parse public key
	block, _ := pem.Decode([]byte(v.config.JWTPublicKey))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ECDSA public key: %w", err)
	}

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pub, nil
	})
}

// validateJWKS validates token using JWKS endpoint
func (v *JWTValidator) validateJWKS(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Get key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("token missing kid header")
		}

		// Fetch JWKS
		resp, err := http.Get(v.config.JWKSEndpoint)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read JWKS response: %w", err)
		}

		// Parse JWKS
		var jwks struct {
			Keys []map[string]interface{} `json:"keys"`
		}
		if err := json.Unmarshal(body, &jwks); err != nil {
			return nil, fmt.Errorf("failed to parse JWKS: %w", err)
		}

		// Find matching key
		for _, key := range jwks.Keys {
			if keyID, ok := key["kid"].(string); ok && keyID == kid {
				// For simplicity, we'll handle RSA keys for now
				// In production, you'd want to use a proper JWKS library
				if kty, ok := key["kty"].(string); ok && kty == "RSA" {
					// Convert JWK to RSA public key
					// This is a simplified version - production should use a proper JWKS library
					return nil, fmt.Errorf("JWKS validation requires additional implementation")
				}
			}
		}

		return nil, fmt.Errorf("no matching key found in JWKS")
	})
}

// validateStandardClaims validates issuer, audience, and expiration
func (v *JWTValidator) validateStandardClaims(claims jwt.MapClaims) error {
	// Validate issuer if configured
	if v.config.JWTIssuer != "" {
		iss, ok := claims["iss"].(string)
		if !ok || iss != v.config.JWTIssuer {
			return fmt.Errorf("invalid issuer: expected %s, got %s", v.config.JWTIssuer, iss)
		}
	}

	// Validate audience if configured
	if v.config.JWTAudience != "" {
		aud, ok := claims["aud"].(string)
		if !ok || aud != v.config.JWTAudience {
			return fmt.Errorf("invalid audience: expected %s, got %s", v.config.JWTAudience, aud)
		}
	}

	// Validate expiration with buffer
	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("missing or invalid exp claim")
	}

	expiresAt := time.Unix(int64(exp), 0)
	buffer := time.Duration(v.config.TokenExpirationBuffer) * time.Second
	if time.Now().After(expiresAt.Add(buffer)) {
		return fmt.Errorf("token has expired")
	}

	return nil
}

// extractUserFromClaims extracts user information from JWT claims
func (v *JWTValidator) extractUserFromClaims(claims jwt.MapClaims) (*models.KotomiUser, error) {
	// Get subject (user ID)
	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return nil, fmt.Errorf("missing or invalid sub claim")
	}

	// Get kotomi_user object
	kotomiUserClaim, ok := claims["kotomi_user"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid kotomi_user claim")
	}

	user := &models.KotomiUser{
		ID: sub,
	}

	// Extract required fields
	if name, ok := kotomiUserClaim["name"].(string); ok {
		user.Name = name
	} else {
		return nil, fmt.Errorf("missing or invalid kotomi_user.name")
	}

	// Extract optional fields
	if email, ok := kotomiUserClaim["email"].(string); ok {
		user.Email = email
	}

	if avatarURL, ok := kotomiUserClaim["avatar_url"].(string); ok {
		user.AvatarURL = avatarURL
	}

	if profileURL, ok := kotomiUserClaim["profile_url"].(string); ok {
		user.ProfileURL = profileURL
	}

	if verified, ok := kotomiUserClaim["verified"].(bool); ok {
		user.Verified = verified
	}

	// Extract roles array
	if roles, ok := kotomiUserClaim["roles"].([]interface{}); ok {
		user.Roles = make([]string, 0, len(roles))
		for _, role := range roles {
			if roleStr, ok := role.(string); ok {
				user.Roles = append(user.Roles, roleStr)
			}
		}
	}

	return user, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) string {
	if authHeader == "" {
		return ""
	}

	// Bearer token format: "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}
