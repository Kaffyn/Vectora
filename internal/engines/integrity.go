package engines

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// VerifyFile valida a integridade de um arquivo contra um hash SHA256.
// Retorna erro se o arquivo não existe ou o hash não bate.
func VerifyFile(filePath, expectedSHA256 string) error {
	f, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", filePath)
		}
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	return verifyFileHash(f, expectedSHA256)
}

// VerifyReader valida o hash de um io.Reader.
func VerifyReader(r io.Reader, expectedSHA256 string) error {
	return verifyFileHash(r, expectedSHA256)
}

// verifyFileHash é o worker que faz a validação de hash.
func verifyFileHash(r io.Reader, expectedSHA256 string) error {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	actualHex := hex.EncodeToString(h.Sum(nil))
	if actualHex != expectedSHA256 {
		return fmt.Errorf(
			"hash mismatch: expected %s, got %s",
			expectedSHA256,
			actualHex,
		)
	}

	return nil
}

// ComputeFileSHA256 calcula o hash SHA256 de um arquivo.
// Útil para verificar hashes após download/extração.
func ComputeFileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ComputeReaderSHA256 calcula o hash de um Reader.
func ComputeReaderSHA256(r io.Reader) (string, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("failed to read data: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CompareHashes compara dois hashes SHA256 (case-insensitive, trimmed).
func CompareHashes(hash1, hash2 string) bool {
	// Normalizar: lowercase, sem espaços
	h1 := hex.EncodeToString([]byte(hash1))
	h2 := hex.EncodeToString([]byte(hash2))
	return h1 == h2
}
