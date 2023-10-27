package dnskey

import (
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/miekg/dns"
)

type DNSKeyVals struct {
	Zone string `json:"zone,omitempty"`
	Base string `json:"base,omitempty"`
	Key  string `json:"key,omitempty"`
	Priv string `json:"priv,omitempty"`
	DS   string `json:"ds,omitempty"`
}

func (d DNSKeyVals) ToMap() map[string]any {
	return map[string]any{
		`base`: d.Base,
		`key`:  d.Key,
		`priv`: d.Priv,
		`ds`:   d.DS,
	}
}

func GenerateDNSKey(zoneName string) *DNSKeyVals {
	// taken from: https://github.com/coredns/coredns-utils/blob/master/coredns-keygen/main.go
	key := &dns.DNSKEY{
		Hdr:       dns.RR_Header{Name: dns.Fqdn(zoneName), Class: dns.ClassINET, Ttl: 3600, Rrtype: dns.TypeDNSKEY},
		Algorithm: dns.ECDSAP256SHA256, Flags: 257, Protocol: 3,
	}
	priv, err := key.Generate(256)
	if err != nil {
		hclog.Default().Error(`dnskey.key.Generate256`, `err`, err)
		return nil
	}

	ds := key.ToDS(dns.SHA256)

	base := fmt.Sprintf("K%s+%03d+%05d", key.Header().Name, key.Algorithm, key.KeyTag())
	// base+".key", []byte(key.String()+"\n")
	// base+".private", []byte(key.PrivateKeyString(priv))
	// base+".ds", []byte(ds.String()+"\n")
	return &DNSKeyVals{
		zoneName,
		base,
		key.String(),
		key.PrivateKeyString(priv),
		ds.String(),
	}
}
