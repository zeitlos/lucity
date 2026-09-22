package hostname

import (
	"crypto/sha256"
	"encoding/hex"
)

const verifyRecordPrefix = "_lucity-verify."

func (c *Client) DNSRecords(workspace, host string, redirect bool) []DNSRecord {
	if c.IsPlatform(host) || c.IsInternal(host) {
		return nil
	}

	zone := registrableDomain(host)

	records := []DNSRecord{
		{
			Type:  TXT,
			Host:  verifyRecordPrefix + zone,
			Value: challenge(workspace, zone),
		},
	}

	if isApex(host) && redirect {
		records = append(records, DNSRecord{
			Type:  A,
			Host:  host,
			Value: c.loadBalancerIP,
		})

		return records
	}

	if isApex(host) {
		records = append(records, DNSRecord{
			Type:  ALIAS,
			Host:  host,
			Value: c.loadBalancerHostname,
		})

		return records
	}

	records = append(records, DNSRecord{
		Type:  CNAME,
		Host:  host,
		Value: c.loadBalancerHostname,
	})

	return records
}

func challenge(workspace, host string) string {
	sum := sha256.Sum256([]byte(workspace + ":" + host))
	return hex.EncodeToString(sum[:16])
}
