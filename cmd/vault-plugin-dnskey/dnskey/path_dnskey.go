package dnskey

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
)

// pathDnskey extends the Vault API with a `/dnskey`
// endpoint for the backend. You can choose whether
// certain attributes should be displayed,
// required, and named. For example, password
// is marked as sensitive and will not be output
// when you read the configuration.
func pathDnskey(b *dnskeyBackend) *framework.Path {
	return &framework.Path{
		Pattern: "dnskey/.*",
		Fields: map[string]*framework.FieldSchema{
			"zone": {
				Type: framework.TypeString,
			},
		},
		Operations: map[logical.Operation]framework.OperationHandler{
			logical.ReadOperation: &framework.PathOperation{
				Callback: b.pathDnskeyRead,
			},
			logical.CreateOperation: &framework.PathOperation{
				Callback: b.pathDnskeyWrite,
			},
			logical.UpdateOperation: &framework.PathOperation{
				Callback: b.pathDnskeyWrite,
			},
			logical.DeleteOperation: &framework.PathOperation{
				Callback: b.pathDnskeyDelete,
			},
		},
		ExistenceCheck:  b.pathDnskeyExistenceCheck,
		HelpSynopsis:    pathConfigHelpSynopsis,
		HelpDescription: pathConfigHelpDescription,
	}
}

// pathDnskeyExistenceCheck verifies if the configuration exists.
func (b *dnskeyBackend) pathDnskeyExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	hclog.Default().Info(`dnskey plugin existence check`, `path`, req.Path, `data`, data.Raw, `schema`, data.Schema)
	out, err := req.Storage.Get(ctx, req.Path)
	if err != nil {
		return false, fmt.Errorf("existence check failed: %w", err)
	}

	return out != nil, nil
}

// pathDnskeyRead reads the configuration and outputs non-sensitive information.
func (b *dnskeyBackend) pathDnskeyRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	hclog.Default().Info(`dnskey plugin read`, `path`, req.Path, `data`, data.Raw, `schema`, data.Schema)
	dnsKey, err := getDnskey(ctx, req.Storage, req.Path)
	if err != nil {
		return nil, err
	}

	return &logical.Response{
		Data: dnsKey.ToMap(),
	}, nil
}

// pathDnskeyWrite updates the configuration for the backend
func (b *dnskeyBackend) pathDnskeyWrite(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	hclog.Default().Info(`dnskey plugin write`, `path`, req.Path, `data`, data.Raw, `schema`, data.Schema)
	dnsKey, err := getDnskey(ctx, req.Storage, req.Path)
	if err != nil {
		return nil, err
	}

	if dnsKey == nil {
		if req.Operation == logical.UpdateOperation {
			return nil, errors.New("dnsKey not found during update operation")
		}
		dnsKey = new(DNSKeyVals)
	}

	zone, ok := data.GetOk("zone")
	if !ok {
		return nil, errors.New("zone is required")
	}

	dnsKey = GenerateDNSKey(zone.(string))
	if dnsKey == nil {
		return nil, errors.New("cannot generate dnsKey")
	}

	entry, err := logical.StorageEntryJSON(req.Path, dnsKey)
	if err != nil {
		return nil, err
	}

	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	// reset the client so the next invocation will pick up the new configuration
	b.reset()

	return &logical.Response{
		Data: dnsKey.ToMap(),
	}, nil
}

// pathDnskeyDelete removes the configuration for the backend
func (b *dnskeyBackend) pathDnskeyDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	hclog.Default().Info(`dnskey plugin delete`, `path`, req.Path, `data`, data.Raw, `schema`, data.Schema)
	err := req.Storage.Delete(ctx, req.Path)

	if err == nil {
		b.reset()
	}

	return nil, err
}

func getDnskey(ctx context.Context, s logical.Storage, path string) (*DNSKeyVals, error) {
	hclog.Default().Info(`dnskey plugin get`, `path`, path)
	entry, err := s.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	if entry == nil {
		return nil, nil
	}

	config := new(DNSKeyVals)
	if err := entry.DecodeJSON(&config); err != nil {
		return nil, fmt.Errorf("error reading root configuration: %w", err)
	}

	// return the config, we are done
	return config, nil
}

// pathConfigHelpSynopsis summarizes the help text for the configuration
const pathConfigHelpSynopsis = `Configure the DNSKey backend.`

// pathConfigHelpDescription describes the help text for the configuration
const pathConfigHelpDescription = `
The DNSKey secret backend requires credentials for managing
DNSKey issued for specific zones.
`
