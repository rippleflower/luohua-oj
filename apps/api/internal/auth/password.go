package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	legacyHashIterations = 120000
	argonTimeCost        = 3
	argonMemoryCost      = 64 * 1024
	argonParallelism     = 4
	hashKeyLen           = 32
	saltLen              = 16
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTimeCost, argonMemoryCost, argonParallelism, hashKeyLen)
	return fmt.Sprintf(
		"argon2id$%d$%d$%d$%s$%s",
		argonTimeCost,
		argonMemoryCost,
		argonParallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func verifyPassword(password string, encoded string) error {
	parts := strings.Split(encoded, "$")
	if len(parts) == 0 {
		return errors.New("invalid password hash")
	}

	switch parts[0] {
	case "argon2id":
		return verifyArgon2Password(password, parts)
	case "sha256":
		return verifyLegacyPassword(password, parts)
	default:
		return errors.New("invalid password hash")
	}
}

func passwordNeedsUpgrade(encoded string) bool {
	return !strings.HasPrefix(encoded, "argon2id$")
}

func verifyArgon2Password(password string, parts []string) error {
	if len(parts) != 6 {
		return errors.New("invalid password hash")
	}

	timeCost, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return errors.New("invalid password hash")
	}
	memoryCost, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil {
		return errors.New("invalid password hash")
	}
	parallelism, err := strconv.ParseUint(parts[3], 10, 8)
	if err != nil {
		return errors.New("invalid password hash")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return err
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return err
	}
	actual := argon2.IDKey([]byte(password), salt, uint32(timeCost), uint32(memoryCost), uint8(parallelism), uint32(len(expected)))
	if len(actual) != len(expected) {
		return errors.New("invalid credentials")
	}
	if subtle.ConstantTimeCompare(actual, expected) != 1 {
		return errors.New("invalid credentials")
	}
	return nil
}

func verifyLegacyPassword(password string, parts []string) error {
	if len(parts) != 4 {
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
	actual := deriveLegacyPasswordHash(password, salt)
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

func deriveLegacyPasswordHash(password string, salt []byte) []byte {
	sum := sha256.Sum256(append(append([]byte{}, salt...), []byte(password)...))
	state := sum[:]
	for i := 1; i < legacyHashIterations; i++ {
		next := sha256.Sum256(append(append([]byte{}, state...), []byte(password)...))
		state = next[:]
	}
	return append([]byte{}, state[:hashKeyLen]...)
}
