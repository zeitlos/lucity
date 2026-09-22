package edge

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/zeitlos/lucity/pkg/bunny"
)

const handshakeTimeout = 10 * time.Second

func (c *Client) ensureUploaded(ctx context.Context, zone *bunny.PullZone, name string, certificate *Certificate) error {
	serial, err := leafSerial(certificate.Certificate)

	if err != nil {
		return err
	}

	if servedSerial(ctx, zone.Name+".b-cdn.net", name) == serial {
		return nil
	}

	return c.api.UploadCertificate(ctx, zone.ID, name, certificate.Certificate, certificate.Key)
}

func leafSerial(certificatePEM []byte) (string, error) {
	block, _ := pem.Decode(certificatePEM)

	if block == nil {
		return "", errors.New("certificate is not PEM encoded")
	}

	leaf, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		return "", err
	}

	return leaf.SerialNumber.Text(16), nil
}

func servedSerial(ctx context.Context, edge, name string) string {
	dialer := tls.Dialer{
		NetDialer: &net.Dialer{Timeout: handshakeTimeout},
		Config:    &tls.Config{ServerName: strings.Replace(name, "*.", "wildcard-probe.", 1), InsecureSkipVerify: true},
	}

	conn, err := dialer.DialContext(ctx, "tcp", edge+":443")

	if err != nil {
		return ""
	}

	defer conn.Close()

	certificates := conn.(*tls.Conn).ConnectionState().PeerCertificates

	if len(certificates) == 0 {
		return ""
	}

	return certificates[0].SerialNumber.Text(16)
}
