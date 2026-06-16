package hostname

import (
	"strings"

	"golang.org/x/net/publicsuffix"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type Client struct {
	workloadDomain    string
	customCNAMETarget string
	customApexIP      string
	gatewayNamespace  string
	k8s               kubernetes.Interface
	dyn               dynamic.Interface
}

func New(workloadDomain, customCNAMETarget, customApexIP, gatewayNamespace string, k8s kubernetes.Interface, dyn dynamic.Interface) *Client {
	return &Client{
		workloadDomain:    workloadDomain,
		customCNAMETarget: customCNAMETarget,
		customApexIP:      customApexIP,
		gatewayNamespace:  gatewayNamespace,
		k8s:               k8s,
		dyn:               dyn,
	}
}

type DNSStatus string

const (
	DNSValid         DNSStatus = "valid"
	DNSPending       DNSStatus = "pending"
	DNSMisconfigured DNSStatus = "misconfigured"
	DNSError         DNSStatus = "error"
)

type TLSStatus string

const (
	TLSNone         TLSStatus = "none"
	TLSProvisioning TLSStatus = "provisioning"
	TLSActive       TLSStatus = "active"
	TLSError        TLSStatus = "error"
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

func (c *Client) IsInternal(host string) bool {
	return strings.HasSuffix(host, ".local")
}

func (c *Client) IsCustom(host string) bool {
	return !c.IsInternal(host) && !c.IsPlatform(host)
}

func registrableDomain(host string) string {
	host = strings.TrimSuffix(host, ".")

	domain, err := publicsuffix.EffectiveTLDPlusOne(host)

	if err != nil {
		return host
	}

	return domain
}

func isApex(host string) bool {
	return registrableDomain(host) == strings.TrimSuffix(host, ".")
}
