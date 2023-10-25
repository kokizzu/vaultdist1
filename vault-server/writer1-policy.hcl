
# Grant 'update' permission on the 'auth/approle/role/<role_name>/secret-id' path for generating a secret id
path "auth/approle/role/reader1/secret-id" {
  capabilities = ["update"]
}

path "auth/approle/role/writer1/secret-id" {
  capabilities = ["update"]
}

path "secret/data/keys1/*" {
  capabilities = ["create","update","read","patch","delete"]
}

path "secret/keys1/*" { # v1
  capabilities = ["create","update","read","patch","delete"]
}

path "secret/metadata/keys1/*" {
  capabilities = ["list"]
}