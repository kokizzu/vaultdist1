
# Vault DNSKEY Store Demo

Demo showing how to use Vault to store, read, and delete DNSKEY records with
the built-in KV engine and with a custom Vault plugin.

## Requirements

- Go 1.26.5 or newer
- Docker Compose for the Vault demo environment

## Build plugin

this will build the plugin

```
make build-plugin
```

## Start Dependencies

this also inject a writer token (the application that will fetch reader token, set policies), also copy plugin to docker and enable it

```
docker-compose up --build
```

fetch reader1 and writer1 token and write it to `/tmp/reader1-secret` and `/tmp/writer1-secret`

```
./fetch-tokens.sh
```

example how to write new dnskey (using `writer1_approle1` and `/tmp/writer1-secret`)

```
go run main.go write test.com
go run main.go writeplugin test.com
```

example how to read the previously stored dnskey (using `reader1_approle1` and `/tmp/reader1-secret`)

```
go run main.go read test.com
go run main.go readplugin test.com
```

benchmark

```
go run main.go benchmark
writeCounter: 3923 (632.22/s), readCounter: 18048 (2908.55/s), err: 3
go run main.go benchmarkplugin
writeCounter: 14545 (450.98/s), readCounter: 75257 (2333.41/s), err: 0
```

NOTE:

- there's [issue with vault](//github.com/hashicorp/vault/issues/23814) (only for default kv engine, plugin version works fine), that it can only store 5000-ish key before starting to get `read: connection reset by peer`, after a while it can write again, then get same error again 
after 5000-ish, repeat.
- `deleteplugin` didn't work, not sure what's the issue, probably because the storage always versioned

## Verification

Run local tests without starting Vault:

```sh
make test
```

Run a dependency vulnerability check:

```sh
make vulncheck
```

Run any one-off command through the Makefile:

```sh
make run CMD='go test ./...'
```

## Maintenance Checklist

- [x] Update the Go runtime directive to Go 1.26.5 for the fixed standard library.
- [x] Refresh Vault, DNS, and helper dependencies.
- [x] Add local DNSKEY generation and plugin lifecycle tests.
- [x] Add Makefile targets for tests, plugin builds, vulnerability checks, and arbitrary commands.
- [x] Run `make test`.
- [x] Run `make vulncheck`; no reachable vulnerabilities were found.


