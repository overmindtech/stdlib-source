package network

import (
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"fmt"
)

// CertToName Returns the name of a cert as a string. This is in the format of:
//
// {Subject.CommonName} (SHA-1: {hash})
func CertToName(cert *x509.Certificate) string {
	sum := sha1.Sum(cert.Raw)

	return fmt.Sprintf(
		"%v (SHA-1: %v)",
		cert.Subject.CommonName,
		hex.EncodeToString(sum[:]),
	)
}
