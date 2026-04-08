package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// MasterKey deriva uma chave de 32 bytes para AES.
// Em produção, isso viria de um KDF (Argon2) baseado em senha ou Keyring do OS.
func getMasterKey() ([]byte, error) {
	// Simplificação para MVP: Usa uma variável de ambiente do sistema ou fallback
	keyStr := os.Getenv("VECTORA_MASTER_KEY")
	if keyStr == "" {
		// Fallback inseguro apenas para desenvolvimento local sem setup
		keyStr = "vectora-dev-master-key-change-me-in-prod!!"
	}
	hash := sha256.Sum256([]byte(keyStr))
	return hash[:], nil
}

func EncryptSecret(plaintext string) (string, error) {
	key, err := getMasterKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptSecret(ciphertextB64 string) (string, error) {
	key, err := getMasterKey()
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
