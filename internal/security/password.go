package security

import (
	"crypto/rand"
	"encoding/base64"
	"golang.org/x/crypto/argon2"
	"log"
)

type PasswordObj struct {
	Hash string
	Salt string
}

func generateSalt() ([]byte, error) {
	salt := make([]byte, 16)

	_, err := rand.Read(salt)
	if err != nil {
		return nil, err
	}
	return salt, nil
}

func GeneratePasswd(pass string) (PasswordObj, error) {
	salt, err := generateSalt()
	if err != nil {

		log.Printf("Error generating salt: %v", err)
		return PasswordObj{}, err
	}
	key := argon2.IDKey([]byte(pass), salt, 1, 64*1024, 4, 32)
	base64Hash := base64.RawStdEncoding.EncodeToString(key)
	base64Salt := base64.RawStdEncoding.EncodeToString(salt)
	return PasswordObj{Hash: base64Hash, Salt: base64Salt}, nil
}
func ComparePasswords(passwd string, passwdDB string, salt string) bool {
	saltByte, _ := base64.RawStdEncoding.DecodeString(salt)
	userPassHash := argon2.IDKey([]byte(passwd), saltByte, 1, 64*1024, 4, 32)
	base64Hash := base64.RawStdEncoding.EncodeToString(userPassHash)
	if base64Hash != passwdDB {
		return false
	}
	return true
}
