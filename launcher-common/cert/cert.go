package cert

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"os"
)

func BytesToCertificate(data []byte) *x509.Certificate {
	cert, _ := x509.ParseCertificate(data)
	return cert
}

func WriteAsPem(data []byte, file *os.File) error {
	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: data,
	})
	_, err := file.Write(pemData)
	return err
}

func ReadFromFile(filePath string) (keys []string, keyToIndex map[string]int, values []*x509.Certificate, err error) {
	var pemData []byte
	pemData, err = os.ReadFile(filePath)
	if err != nil {
		return
	}

	keys = make([]string, 0)
	values = make([]*x509.Certificate, 0)
	keyToIndex = make(map[string]int)
	var cert *x509.Certificate
	for {
		var block *pem.Block
		block, pemData = pem.Decode(pemData)
		if block == nil {
			break
		}

		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return
		}

		hash := sha256.Sum256(cert.Raw)
		fingerprint := hex.EncodeToString(hash[:])

		keys = append(keys, fingerprint)
		values = append(values, cert)
		keyToIndex[fingerprint] = len(keys) - 1

		if len(pemData) == 0 {
			break
		}
	}

	return
}
