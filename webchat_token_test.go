package findai

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGenerateWebchatVisitorToken(t *testing.T) {
	secret := "test_secret_key_123"
	visitorID := "user_abc123"
	ttl := 2 * time.Hour

	token, err := GenerateWebchatVisitorToken(secret, visitorID, ttl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected 3 JWT parts, got %d", len(parts))
	}

	// Verify header
	headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	var header map[string]string
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		t.Fatalf("unmarshal header: %v", err)
	}
	if header["alg"] != "HS256" || header["typ"] != "JWT" {
		t.Errorf("unexpected header: %v", header)
	}

	// Verify claims
	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("decode claims: %v", err)
	}
	var claims map[string]any
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		t.Fatalf("unmarshal claims: %v", err)
	}
	if claims["visitor_id"] != visitorID {
		t.Errorf("visitor_id = %v, want %v", claims["visitor_id"], visitorID)
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatalf("exp not a number: %v", claims["exp"])
	}
	expectedExp := time.Now().Add(ttl).Unix()
	if int64(exp) < expectedExp-5 || int64(exp) > expectedExp+5 {
		t.Errorf("exp %v not within expected range ~%v", int64(exp), expectedExp)
	}

	// Verify signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if parts[2] != expectedSig {
		t.Errorf("signature mismatch")
	}
}

func TestGenerateWebchatVisitorToken_Validation(t *testing.T) {
	_, err := GenerateWebchatVisitorToken("", "visitor", time.Hour)
	if err == nil {
		t.Error("expected error for empty secret")
	}

	_, err = GenerateWebchatVisitorToken("secret", "", time.Hour)
	if err == nil {
		t.Error("expected error for empty visitor_id")
	}
}

func TestGenerateWebchatVisitorToken_DefaultTTL(t *testing.T) {
	token, err := GenerateWebchatVisitorToken("secret", "v1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parts := strings.Split(token, ".")
	claimsJSON, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var claims map[string]any
	json.Unmarshal(claimsJSON, &claims)

	exp := int64(claims["exp"].(float64))
	expectedExp := time.Now().Add(time.Hour).Unix()
	if exp < expectedExp-5 || exp > expectedExp+5 {
		t.Errorf("default TTL exp %v not within expected range ~%v", exp, expectedExp)
	}
}
