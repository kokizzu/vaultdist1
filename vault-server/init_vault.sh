#!/bin/sh

set -e

export VAULT_ADDR='http://127.0.0.1:8200'
export VAULT_FORMAT='json'

ls -al .

# Spawn a new process for the development Vault server and wait for it to come online
# ref: https://www.vaultproject.io/docs/concepts/dev-server
vault server -log-level=debug -dev -dev-listen-address="0.0.0.0:8200" -dev-plugin-dir=/plugins &
sleep 1s

# authenticate container's local Vault CLI
# ref: https://www.vaultproject.io/docs/commands/login
vault login -no-print "${VAULT_DEV_ROOT_TOKEN_ID}"

# add policy
# ref: https://www.vaultproject.io/docs/concepts/policies
vault policy write reader1-policy /vault/config/reader1-policy.hcl
vault policy write writer1-policy /vault/config/writer1-policy.hcl

# enable AppRole auth method
# ref: https://www.vaultproject.io/docs/auth/approle
vault auth enable approle

# configure AppRole
# ref: https://www.vaultproject.io/api/auth/approle#parameters
vault write auth/approle/role/reader1 \
    token_policies=reader1-policy \
    token_num_uses=0 \
    secret_id_ttl="32d" \
    token_ttl="32d" \
    token_max_ttl="32d"

# overwrite our role id
vault write auth/approle/role/reader1/role-id role_id="${READER1_APPROLE_ID}"

vault write auth/approle/role/writer1 \
    token_policies=writer1-policy \
    token_num_uses=0 \
    secret_id_ttl="32d" \
    token_ttl="32d" \
    token_max_ttl="32d"

vault write auth/approle/role/writer1/role-id role_id="${WRITER1_APPROLE_ID}"

# ref: https://www.vaultproject.io/docs/commands/token/create
vault token create \
    -id="${WRITER1_TOKEN}" \
    -policy=writer1-policy \
    -ttl="32d"

vault secrets enable -path=mysecret vault-plugin-dnskey

vault secrets list

# keep container alive
tail -f /dev/null & trap 'kill %1' TERM ; wait