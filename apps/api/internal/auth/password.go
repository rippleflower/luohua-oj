package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

const (
	hashIterations = 120000
	hashKeyLen     = 32
	saltLen        = 16
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := derivePasswordHash(password, salt)
	return fmt.Sprintf(
		"sha256$%d$%s$%s",
		hashIterations,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func verifyPassword(password string, encoded string) error {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "sha256" {
		return errors.New("invalid password hash")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return err
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return err
	}
	actual := derivePasswordHash(password, salt)
	if len(actual) != len(expected) {
		return errors.New("invalid credentials")
	}
	if subtle.ConstantTimeCompare(actual, expected) != 1 {
		return errors.New("invalid credentials")
	}
	return nil
}

func newSessionToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(buf)
	hash := sha256.Sum256([]byte(token))
	return token, base64.RawURLEncoding.EncodeToString(hash[:]), nil
}

func derivePasswordHash(password string, salt []byte) []byte {
	sum := sha256.Sum256(append(append([]byte{}, salt...), []byte(password)...))
	state := sum[:]
	for i := 1; i < hashIterations; i++ {
		next := sha256.Sum256(append(append([]byte{}, state...), []byte(password)...))
		state = next[:]
	}
	return append([]byte{}, state[:hashKeyLen]...)
}
