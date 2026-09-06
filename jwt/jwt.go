package jwt

import (
	"encoding/json"
	"fmt"
	"time"

	gjwt "github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Secret     string
	Expiration time.Duration
	Issuer     string
	Audience   string
}
type Client struct{ config Config }

func New(cfg Config) (*Client, error) {
	if cfg.Secret == "" {
		return nil, fmt.Errorf("jwt secret is not configured")
	}
	return &Client{config: cfg}, nil
}

func (c *Client) Sign(claims any) (string, error) {
	if c == nil || c.config.Secret == "" {
		return "", fmt.Errorf("jwt secret is not configured")
	}
	data, err := structToMap(claims)
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	data["iat"] = now.Unix()
	if c.config.Expiration > 0 {
		data["exp"] = now.Add(c.config.Expiration).Unix()
	}
	if c.config.Issuer != "" {
		data["iss"] = c.config.Issuer
	}
	if c.config.Audience != "" {
		data["aud"] = c.config.Audience
	}
	return gjwt.NewWithClaims(gjwt.SigningMethodHS256, gjwt.MapClaims(data)).SignedString([]byte(c.config.Secret))
}

func (c *Client) Verify(tokenString string, result any) error {
	if c == nil || c.config.Secret == "" {
		return fmt.Errorf("jwt secret is not configured")
	}
	claims := gjwt.MapClaims{}
	options := []gjwt.ParserOption{gjwt.WithValidMethods([]string{gjwt.SigningMethodHS256.Alg()})}
	if c.config.Issuer != "" {
		options = append(options, gjwt.WithIssuer(c.config.Issuer))
	}
	if c.config.Audience != "" {
		options = append(options, gjwt.WithAudience(c.config.Audience))
	}
	token, err := gjwt.ParseWithClaims(tokenString, claims, func(token *gjwt.Token) (any, error) {
		if token.Method != gjwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(c.config.Secret), nil
	}, options...)
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("invalid token")
	}
	data, err := json.Marshal(claims)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}

func structToMap(value any) (map[string]any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("claims must encode to a JSON object")
	}
	return result, nil
}

var defaultClient *Client

func Configure(cfg Config) { defaultClient, _ = New(cfg) }
func Sign[T any](claims T) (string, error) {
	if defaultClient == nil {
		return "", fmt.Errorf("jwt secret is not configured")
	}
	return defaultClient.Sign(claims)
}
func Verify[T any](tokenString string) (T, error) {
	var result T
	if defaultClient == nil {
		return result, fmt.Errorf("jwt secret is not configured")
	}
	return result, defaultClient.Verify(tokenString, &result)
}
