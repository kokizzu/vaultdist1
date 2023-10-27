package dnskey

import (
	"context"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// Factory returns a new backend as logical.Backend
func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := backend()
	hclog.Default().Info(`dnskey plugin setup`)
	if err := b.Setup(ctx, conf); err != nil {
		return nil, err
	}
	return b, nil
}

type dnskeyBackend struct {
	*framework.Backend
	lock sync.RWMutex
}

// backend defines the target API backend
// for Vault. It must include each path
// and the secrets it will store.
func backend() *dnskeyBackend {
	var b = dnskeyBackend{}

	b.Backend = &framework.Backend{
		Help: strings.TrimSpace(backendHelp),
		PathsSpecial: &logical.Paths{
			LocalStorage: []string{
				// WAL stands for Write-Ahead-Log, which is used for Vault replication
				framework.WALPrefix,
			},
			SealWrapStorage: []string{
				"dnskey",
				"secret/*",
			},
		},
		Paths: framework.PathAppend(
			[]*framework.Path{
				pathDnskey(&b),
			},
		),
		Secrets:     []*framework.Secret{},
		BackendType: logical.TypeLogical,
		Invalidate:  b.invalidate,
	}
	return &b
}

// reset clears any client configuration for a new
// backend to be configured
func (b *dnskeyBackend) reset() {
	b.lock.Lock()
	defer b.lock.Unlock()
}

// invalidate clears an existing client configuration in
// the backend
func (b *dnskeyBackend) invalidate(ctx context.Context, key string) {
	if key == "config" {
		b.reset()
	}
}

var backendHelp = `
The DNSKEY plugin allows Vault to manage DNSSEC keys for a DNS zone.
`
