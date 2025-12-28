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

type keyPair struct {
	CertPEMFile       string
	PrivateKeyPEMFile string
}

func genSertKey(path string) (keyPair, error) {
	errText := "unable to generate cert sequence err: "
	// создаём шаблон сертификата
	cert := &x509.Certificate{
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

	// создаём новый приватный RSA-ключ длиной 4096 бит
	// обратите внимание, что для генерации ключа и сертификата
	// используется rand.Reader в качестве источника случайных данных
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return keyPair{}, fmt.Errorf(errText+" %w", err)
	}

	// создаём сертификат x.509
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return keyPair{}, fmt.Errorf(errText+" %w", err)
	}

	// кодируем сертификат и ключ в формате PEM, который
	// используется для хранения и обмена криптографическими ключами
	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return keyPair{}, fmt.Errorf(errText+" %w", err)
	}

	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return keyPair{}, fmt.Errorf(errText+" %w", err)
	}

	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return keyPair{}, fmt.Errorf(errText+" %w", err)
	}

	if err = os.WriteFile(filepath.Join(path, "cert.pem"), certPEM.Bytes(), 0644); err != nil {
		return keyPair{}, fmt.Errorf(errText+" %w", err)
	}

	if err = os.WriteFile(filepath.Join(path, "private.pem"), privateKeyPEM.Bytes(), 0644); err != nil {
		return keyPair{}, fmt.Errorf(errText+" %w", err)
	}
	return keyPair{
		CertPEMFile:       filepath.Join(path, "cert.pem"),
		PrivateKeyPEMFile: filepath.Join(path, "private.pem"),
	}, nil
}

func GetSertKey(path string) (keyPair, error) {
	certPath := filepath.Join(path, "cert.pem")
	privateKeyPath := filepath.Join(path, "private.pem")

	if !fileExists(certPath) || !fileExists(privateKeyPath) {
		return genSertKey(path)
	}
	return keyPair{
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
