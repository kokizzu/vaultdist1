
# Demo how to use vault as store for DNSKey

how to use vault to store (write/read/delete) DNSKey

## Start Dependencies

this also inject a writer token (the application that will fetch reader token, set policies)

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
```

example how to read the previously stored dnskey (using `reader1_approle1` and `/tmp/reader1-secret`)

```
go run main.go read test.com
```

benchmark

```
go run main.go benchmark
writeCounter: 3923 (632.22/s), readCounter: 18048 (2908.55/s), err: 3
```

there's issue with vault, that it can only store 5000-ish key before starting to get `read: connection reset by peer`, after a while it can write again, then get same error again after 5000-ish, repeat.
