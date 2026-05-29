package hostname

import (
	"strings"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Client struct {
	workloadDomain    string
	customCNAMETarget string
	customApexIP      string
	k8s               kubernetes.Interface
	dyn               dynamic.Interface
}

func New(workloadDomain, customCNAMETarget, customApexIP string, k8s kubernetes.Interface, dyn dynamic.Interface) *Client {
	return &Client{
		workloadDomain:    workloadDomain,
		customCNAMETarget: customCNAMETarget,
		customApexIP:      customApexIP,
		k8s:               k8s,
		dyn:               dyn,
	}
}

type Status struct {
	DNS DNSState
	TLS TLSState
}

type DNSState string

const (
	DNSValid         DNSState = "valid"
	DNSPending       DNSState = "pending"
	DNSMisconfigured DNSState = "misconfigured"
	DNSError         DNSState = "error"
)

type TLSState string

const (
	TLSNone         TLSState = "none"
	TLSProvisioning TLSState = "provisioning"
	TLSActive       TLSState = "active"
	TLSError        TLSState = "error"
)

type DNSRecord struct {
	Type  RecordType
	Host  string
	Value string
}

type RecordType string

const (
	TXT   RecordType = "TXT"
	CNAME RecordType = "CNAME"
	A     RecordType = "A"
)

func (c *Client) IsPlatform(host string) bool {
	return host == c.workloadDomain || strings.HasSuffix(host, "."+c.workloadDomain)
}

func isApex(host string) bool {
	return strings.Count(host, ".") <= 1
}
