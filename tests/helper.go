package testhelpers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"errors"
	"log"
	"math/big"
	"net"
	"os"
	"time"
)

func CreateCertificates(privateFile, publicFile string) {

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatal(err)
	}

	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(privateFile, privateKeyPEM.Bytes(), 0600)
	if err != nil {
		log.Fatal(err)
	}
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"Yandex.Praktikum"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatal(err)
	}

	var publicKeyPEM bytes.Buffer
	err = pem.Encode(&publicKeyPEM, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: certBytes,
	})
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(publicFile, publicKeyPEM.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}

}

func CreateConfigFile(configFile string, config any) (err error) {
	var file *os.File
	file, err = os.Create(configFile)
	if err != nil {
		return
	}
	defer func() {
		err = errors.Join(err, file.Close())
	}()
	var b []byte
	switch c := config.(type) {
	case []byte:
		b = c
	case string:
		b = []byte(c)
	default:
		b, err = json.Marshal(config)

	}
	_, err = file.Write(b)
	return
}
