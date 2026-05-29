package hostname

import (
	"crypto/sha256"
	"encoding/hex"
)

const verifyRecordPrefix = "_lucity-verify."

func (c *Client) DNSRecords(workspace, host string) []DNSRecord {
	if c.IsPlatform(host) {
		return nil
	}

	records := []DNSRecord{
		{
			Type:  TXT,
			Host:  verifyRecordPrefix + host,
			Value: challenge(workspace, host),
		},
	}

	if isApex(host) {
		records = append(records, DNSRecord{
			Type:  A,
			Host:  host,
			Value: c.customApexIP,
		})

		return records
	}

	records = append(records, DNSRecord{
		Type:  CNAME,
		Host:  host,
		Value: c.customCNAMETarget,
	})

	return records
}

func challenge(workspace, host string) string {
	sum := sha256.Sum256([]byte(workspace + ":" + host))
	return hex.EncodeToString(sum[:16])
}
