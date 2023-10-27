
# Demo how to use vault as store for DNSKey

how to use vault to store (write/read/delete) DNSKey

## Build plugin

this will build the plugin

```
# build the plugin
mkdir vault-server/plugins
CGO_ENABLED=0 go build -o vault-server/plugins/vault-plugin-dnskey cmd/vault-plugin-dnskey/main.go
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

- there's [issue with vault](//github.com/hashicorp/vault/issues/23814), that it can only store 5000-ish key before starting to get `read: connection reset by peer`, after a while it can write again, then get same error again 
after 5000-ish, repeat.
- `deleteplugin` didn't work, not sure what's the issue



