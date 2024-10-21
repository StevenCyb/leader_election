package internal

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func GenerateCaCert(t *testing.T) (caPrivateKeyPEM, caCertificatePEM string) {
	t.Helper()

	caPrivateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err, "failed to generate RSA private key for CA")

	caCertTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "CA Root",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caCertificateDER, err := x509.CreateCertificate(rand.Reader, &caCertTemplate, &caCertTemplate, &caPrivateKey.PublicKey, caPrivateKey)
	require.NoError(t, err, "failed to create CA X.509 certificate")

	caPrivateKeyBytes, err := x509.MarshalPKCS8PrivateKey(caPrivateKey)
	require.NoError(t, err, "failed to marshal CA private key to PKCS#8")

	caPrivateKeyPEMBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: caPrivateKeyBytes}
	caCertificatePEMBlock := &pem.Block{Type: "CERTIFICATE", Bytes: caCertificateDER}

	caPrivateKeyPEM = string(pem.EncodeToMemory(caPrivateKeyPEMBlock))
	caCertificatePEM = string(pem.EncodeToMemory(caCertificatePEMBlock))

	return caPrivateKeyPEM, caCertificatePEM
}

func GenerateCert(t *testing.T, commonName string, isClient bool, caPrivateKeyPEM, caCertificatePEM string) (privateKeyPEM, certificatePEM string) {
	t.Helper()

	caPrivateKeyBlock, _ := pem.Decode([]byte(caPrivateKeyPEM))
	require.NotNil(t, caPrivateKeyBlock, "failed to decode CA private key PEM")

	caPrivateKey, err := x509.ParsePKCS8PrivateKey(caPrivateKeyBlock.Bytes)
	require.NoError(t, err, "failed to parse CA private key")

	caCertificateBlock, _ := pem.Decode([]byte(caCertificatePEM))
	require.NotNil(t, caCertificateBlock, "failed to decode CA certificate PEM")

	caCertificate, err := x509.ParseCertificate(caCertificateBlock.Bytes)
	require.NoError(t, err, "failed to parse CA certificate")

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err, "failed to generate RSA private key")

	extKeyUsages := []x509.ExtKeyUsage{}
	if isClient {
		extKeyUsages = append(extKeyUsages, x509.ExtKeyUsageClientAuth)
	} else {
		extKeyUsages = append(extKeyUsages, x509.ExtKeyUsageServerAuth)
	}

	certTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           extKeyUsages,
		BasicConstraintsValid: true,
		DNSNames:              []string{commonName},
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, &certTemplate, caCertificate, &privateKey.PublicKey, caPrivateKey.(*rsa.PrivateKey))
	require.NoError(t, err, "failed to create X.509 certificate")

	privateKeyPEMBlock := &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	certificatePEMBlock := &pem.Block{Type: "CERTIFICATE", Bytes: certificateDER}

	privateKeyPEM = string(pem.EncodeToMemory(privateKeyPEMBlock))
	certificatePEM = string(pem.EncodeToMemory(certificatePEMBlock))

	return privateKeyPEM, certificatePEM
}
