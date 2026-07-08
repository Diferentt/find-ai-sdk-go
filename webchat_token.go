package findai

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// GenerateWebchatVisitorToken creates a signed JWT (HS256) for webchat
// auth_mode="signed". The token carries {visitor_id, exp} and is signed
// with the connection's webhook_secret.
func GenerateWebchatVisitorToken(webhookSecret, visitorID string, ttl time.Duration) (string, error) {
	if webhookSecret == "" {
		return "", fmt.Errorf("findai: webhook_secret is required")
	}
	if visitorID == "" {
		return "", fmt.Errorf("findai: visitor_id is required")
	}
	if ttl <= 0 {
		ttl = time.Hour
	}

	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	claims := map[string]any{
		"visitor_id": visitorID,
		"exp":        time.Now().Add(ttl).Unix(),
	}

	h, err := encodeSegment(header)
	if err != nil {
		return "", err
	}
	p, err := encodeSegment(claims)
	if err != nil {
		return "", err
	}

	signingInput := h + "." + p
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	mac.Write([]byte(signingInput))
	sig := base64URLEncode(mac.Sum(nil))

	return signingInput + "." + sig, nil
}

func encodeSegment(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64URLEncode(b), nil
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}
