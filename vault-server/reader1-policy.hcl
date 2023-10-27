
# This section grants access for the app
path "secret/data/keys1/*" {
  capabilities = ["read"]
}

path "secret/keys1/*" { # v1
  capabilities = ["read"]
}

path "mysecret/*" {
  capabilities = ["read"]
}