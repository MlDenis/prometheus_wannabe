package encrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

var MetricsEncryptor *Encryptor

type Encryptor struct {
	openKey *rsa.PublicKey
}

func InitEncryptor(file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("READ FILE ERR: %w", err)
	}

	keyBlock, _ := pem.Decode(b)
	if keyBlock == nil {
		return fmt.Errorf("EMPTY KEY: %w", err)
	}

	pubKey, err := x509.ParsePKCS1PublicKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("PARSE KEY ERR: %w", err)
	}

	MetricsEncryptor = &Encryptor{
		openKey: pubKey,
	}

	return nil
}

func (m *Encryptor) Encrypt(message []byte) ([]byte, error) {
	hash := sha512.New()
	random := rand.Reader

	step := m.openKey.Size() - 2*hash.Size() - 2
	var encryptedBytes []byte

	for start := 0; start < len(message); start += step {
		finish := start + step
		if finish > len(message) {
			finish = len(message)
		}

		encryptedBlockBytes, err := rsa.EncryptOAEP(hash, random, m.openKey, message[start:finish], nil)
		if err != nil {
			return nil, fmt.Errorf("ENCRYPT MESSAGE PART ERROR: %w", err)
		}

		encryptedBytes = append(encryptedBytes, encryptedBlockBytes...)

	}

	return encryptedBytes, nil
}
