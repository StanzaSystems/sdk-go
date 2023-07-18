package ca

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"

	"github.com/StanzaSystems/sdk-go/logging"
)

func AWSRootCAs(path string) *x509.CertPool {
	certPool := x509.NewCertPool()
	for _, pem := range []string{
		"AmazonRootCA1.pem",
		"AmazonRootCA2.pem",
		"AmazonRootCA3.pem",
		"AmazonRootCA4.pem",
	} {
		cert, err := os.ReadFile(filepath.Join(path, pem))
		if err != nil {
			logging.Error(err)
			return nil
		}
		if !certPool.AppendCertsFromPEM(cert) {
			logging.Error(fmt.Errorf("failed to read file: %s", pem))
			return nil
		}
	}
	logging.Debug("successfully loaded AWS Trust Services CA certs")
	return certPool
}
