package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"os"
)

func ReadRSAPrivateKeyFromFile(path string) (*rsa.PrivateKey, error) {
	privKeyContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(privKeyContent)
	if block == nil {
		return nil, err
	}
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return key, nil
	case "PRIVATE KEY":
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		key, ok := keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, err
		}
		return key, nil
	default:
		return nil, errors.New("unsupported key type")
	}
}

func ReadRSAPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	pubKeyContent, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(pubKeyContent)
	if block == nil {
		return nil, err
	}
	switch block.Type {
	case "PUBLIC KEY":
		keyInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		key, ok := keyInterface.(*rsa.PublicKey)
		if !ok {
			return nil, err
		}
		return key, nil
	case "RSA PUBLIC KEY":
		key, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		return key, nil
	default:
		return nil, errors.New("unsupported key type")
	}
}

func DecryptRSA(input []byte, key *rsa.PrivateKey) ([]byte, error) {
	decryptedBytes, err := rsa.DecryptPKCS1v15(nil, key, input)
	if err != nil {
		return nil, err
	}
	return decryptedBytes, nil
}

func EncryptRSA(input []byte, key *rsa.PublicKey) ([]byte, error) {
	encryptedBytes, err := rsa.EncryptPKCS1v15(rand.Reader, key, input)
	if err != nil {
		return nil, err
	}
	return encryptedBytes, nil
}

func CreateAESSymmetricKey() ([]byte, error) {
	// aes-256, 32 bits
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func EncryptWithAESKey(input []byte, key []byte) (encrypted []byte, nonce []byte, err error) {
	cBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	gcm, err := cipher.NewGCM(cBlock)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	encrypted = gcm.Seal(nil, nonce, input, nil)

	return encrypted, nonce, nil
}

func DecryptWithAESKey(encrypted []byte, key []byte, nonce []byte) (decrypted []byte, err error) {
	cBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(cBlock)
	if err != nil {
		return nil, err
	}
	decrypted, err = gcm.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}
