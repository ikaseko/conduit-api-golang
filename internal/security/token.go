package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/o1egl/paseto"
)

var (
	PasetoFooter       = " "
	pasetoSymmetricKey []byte
)

func Init(key string) error {
	if len(key) != 32 {
		return errors.New("symmetric key must be exactly 32 bytes long")
	}
	pasetoSymmetricKey = []byte(key)
	return nil
}

func GenerateToken(userUID string) string {

	now := time.Now()
	exp := now.Add(time.Hour * 24 * 7)
	//exp := now.Add(time.Second * 10)
	nbt := now
	jsonToken := paseto.JSONToken{
		Audience:   "test",
		Issuer:     "test_service",
		Jti:        GenerateJTI(),
		IssuedAt:   now,
		NotBefore:  nbt,
		Expiration: exp,
		Subject:    "test_subject",
	}
	jsonToken.Set("uid", userUID)
	encrypt, _ := paseto.NewV2().Encrypt(pasetoSymmetricKey, jsonToken, PasetoFooter)
	return encrypt

}

func DecodeToken(token string) (string, error) {
	var newJsonToken paseto.JSONToken
	var newFooter string
	err := paseto.NewV2().Decrypt(token, pasetoSymmetricKey, &newJsonToken, &newFooter)
	if err != nil {
		return "", err
	}
	if newJsonToken.Expiration.Before(time.Now()) {
		return "", errors.New("Token expired")
	}
	return newJsonToken.Get("uid"), nil
}

func GenerateJTI() string {
	bytes := make([]byte, 16) // 16 байт = 128 бит
	rand.Read(bytes)

	return hex.EncodeToString(bytes)
}
