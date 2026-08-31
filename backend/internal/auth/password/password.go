package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	version    = "v=19"
	memory     = 64 * 1024
	iterations = 3
	parallel   = 1
	saltBytes  = 16
	hashBytes  = 32
)

var ErrInvalidHash = errors.New("invalid password hash")

func Hash(plain string) (string, error) {
	if len(plain) < 12 {
		return "", errors.New("password must contain at least 12 characters")
	}
	salt := make([]byte, saltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}
	derived := argon2.IDKey([]byte(plain), salt, iterations, memory, parallel, hashBytes)
	encode := base64.RawStdEncoding.EncodeToString
	return fmt.Sprintf("$argon2id$%s$m=%d,t=%d,p=%d$%s$%s", version, memory, iterations, parallel, encode(salt), encode(derived)), nil
}

func Verify(encoded, plain string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != version {
		return false, ErrInvalidHash
	}
	var memoryCost, timeCost, threads uint32
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memoryCost, &timeCost, &threads); err != nil {
		return false, ErrInvalidHash
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 {
		return false, ErrInvalidHash
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expected) == 0 {
		return false, ErrInvalidHash
	}
	actual := argon2.IDKey([]byte(plain), salt, timeCost, memoryCost, uint8(threads), uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1, nil
}
