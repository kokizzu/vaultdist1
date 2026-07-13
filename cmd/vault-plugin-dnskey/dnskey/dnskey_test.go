package dnskey

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/vault/sdk/logical"
)

func TestGenerateDNSKey(t *testing.T) {
	dnsKey := GenerateDNSKey("example.com")
	if dnsKey == nil {
		t.Fatal("GenerateDNSKey returned nil")
	}
	if dnsKey.Zone != "example.com" {
		t.Fatalf("Zone = %q, want example.com", dnsKey.Zone)
	}
	for name, value := range dnsKey.ToMap() {
		if value == "" {
			t.Fatalf("%s is empty", name)
		}
	}
	if !strings.Contains(dnsKey.Key, "DNSKEY") {
		t.Fatalf("Key = %q, want DNSKEY record", dnsKey.Key)
	}
}

func TestBackendDNSKeyLifecycle(t *testing.T) {
	ctx := context.Background()
	storage := new(logical.InmemStorage)
	backend, err := Factory(ctx, &logical.BackendConfig{StorageView: storage})
	if err != nil {
		t.Fatalf("Factory returned error: %v", err)
	}

	writeResp, err := backend.HandleRequest(ctx, &logical.Request{
		Operation: logical.CreateOperation,
		Path:      "dnskey/example.com",
		Storage:   storage,
		Data: map[string]any{
			"zone": "example.com",
		},
	})
	if err != nil {
		t.Fatalf("write returned error: %v", err)
	}
	if writeResp == nil || writeResp.Data["key"] == "" {
		t.Fatalf("write response missing key: %#v", writeResp)
	}

	readResp, err := backend.HandleRequest(ctx, &logical.Request{
		Operation: logical.ReadOperation,
		Path:      "dnskey/example.com",
		Storage:   storage,
	})
	if err != nil {
		t.Fatalf("read returned error: %v", err)
	}
	if readResp == nil || readResp.Data["ds"] == "" {
		t.Fatalf("read response missing ds: %#v", readResp)
	}

	_, err = backend.HandleRequest(ctx, &logical.Request{
		Operation: logical.DeleteOperation,
		Path:      "dnskey/example.com",
		Storage:   storage,
	})
	if err != nil {
		t.Fatalf("delete returned error: %v", err)
	}

	entry, err := storage.Get(ctx, "dnskey/example.com")
	if err != nil {
		t.Fatalf("storage get returned error: %v", err)
	}
	if entry != nil {
		t.Fatalf("storage entry still exists after delete")
	}
}
