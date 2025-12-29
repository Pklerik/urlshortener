// Package criptography provides functions for generating and managing TLS certificates and keys.
package criptography

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

var (
	cert = &x509.Certificate{
		// указываем уникальный номер сертификата
		SerialNumber: big.NewInt(1658),
		// заполняем базовую информацию о владельце сертификата
		Subject: pkix.Name{
			Organization: []string{"Pavel.Budkov"},
			Country:      []string{"RU"},
		},
		// разрешаем использование сертификата для 127.0.0.1 и ::1
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		// сертификат верен, начиная со времени создания
		NotBefore: time.Now(),
		// время жизни сертификата — 10 лет
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		// устанавливаем использование ключа для цифровой подписи,
		// а также клиентской и серверной авторизации
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}
)

// KeyPair holds the file paths for the certificate and private key.
type KeyPair struct {
	CertPEMFile       string
	PrivateKeyPEMFile string
}

// GetSertKey checks if certificate and private key files exist in the specified path.
// If they do not exist, it generates new ones.
// It returns the paths to the certificate and private key files.
func GetSertKey(path string) (KeyPair, error) {
	certPath := filepath.Join(path, "cert.pem")
	privateKeyPath := filepath.Join(path, "private.pem")

	if !fileExists(certPath) || !fileExists(privateKeyPath) {
		return genSertKey(path)
	}

	return KeyPair{
		CertPEMFile:       certPath,
		PrivateKeyPEMFile: privateKeyPath,
	}, nil
}

// fileExists checks if a file or directory exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // File exists, no error
	}
	// Check if the error is specifically because the file does not exist
	if errors.Is(err, os.ErrNotExist) {
		return false
	}
	// Other errors occurred (e.g., permission denied, invalid path)
	// In most cases, you might still consider the file as "not existing" for a simple check,
	// but it's important to handle or log the underlying error if necessary.
	return false
}

func genSertKey(path string) (KeyPair, error) {
	// создаём шаблон сертификата

	// создаём новый приватный RSA-ключ длиной 4096 бит
	// обратите внимание, что для генерации ключа и сертификата
	// используется rand.Reader в качестве источника случайных данных
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return KeyPair{}, fmt.Errorf("genSertKey: %w", err)
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return KeyPair{}, fmt.Errorf("genSertKey: %w", err)
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var certPEM bytes.Buffer

	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return KeyPair{}, fmt.Errorf("genSertKey: %w", err)
	}

	var privateKeyPEM bytes.Buffer

	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return KeyPair{}, fmt.Errorf("genSertKey: %w", err)
	}

	err = os.MkdirAll(path, 0750)
	if err != nil {
		return KeyPair{}, fmt.Errorf("genSertKey: %w", err)
	}

	if err = os.WriteFile(filepath.Join(path, "cert.pem"), certPEM.Bytes(), 0600); err != nil {
		return KeyPair{}, fmt.Errorf("genSertKey: %w", err)
	}

	if err = os.WriteFile(filepath.Join(path, "private.pem"), privateKeyPEM.Bytes(), 0600); err != nil {
		return KeyPair{}, fmt.Errorf("genSertKey: %w", err)
	}

	return KeyPair{
		CertPEMFile:       filepath.Join(path, "cert.pem"),
		PrivateKeyPEMFile: filepath.Join(path, "private.pem"),
	}, nil
}
