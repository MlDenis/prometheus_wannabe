package encrypt

import (
	"fmt"
	"os"

	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/pem"
)

var MetricsDecryptor *Decryptor

type Decryptor struct {
	privateKey *rsa.PrivateKey
}

func InitializeDecryptor(file string) error {

	b, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("PRIVATE KEY READ FROM FILE ERR: %w", err)
	}

	keyBlock, _ := pem.Decode(b)
	if keyBlock == nil {
		return fmt.Errorf("PRIVATE KEY BLOB ERR: %w", err)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return fmt.Errorf("PRIVATE KEY PARSE ERR: %w", err)
	}

	MetricsDecryptor = &Decryptor{
		privateKey: privateKey,
	}

	return nil
}

func (m *Decryptor) Decrypt(message []byte) ([]byte, error) {
	msgLen := len(message)
	hash := sha512.New()
	random := rand.Reader

	step := m.privateKey.Size()
	var decryptedBytes []byte

	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		decryptedBlockBytes, err := rsa.DecryptOAEP(hash, random, m.privateKey, message[start:finish], nil)
		if err != nil {
			return nil, fmt.Errorf("MESSAGE PART DECRYPT ERR: %w", err)
		}

		decryptedBytes = append(decryptedBytes, decryptedBlockBytes...)
	}

	return decryptedBytes, nil
}
