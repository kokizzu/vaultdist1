
# this script login as writer1 and generate /tmp/reader1-secret and /tmp/writer1-secret

WRITER1_TOKEN=`cat docker-compose.yaml | grep WRITER1_TOKEN | cut -d':' -f2 | xargs echo -n`
VAULT_ADDRESS="127.0.0.1:8200"

echo "'${WRITER1_TOKEN}'"

# retrieve secret for appsecret with writer1 can load the /tmp/reader1-secret
curl -v \
   --request POST \
   --header "X-Vault-Token: ${WRITER1_TOKEN}" \
      "${VAULT_ADDRESS}/v1/auth/approle/role/reader1/secret-id" | tee /tmp/reader1-debug

# if using wrapping token, it the secret id file can only be used once
#   --header "X-Vault-Wrap-TTL: 32d" \
#cat /tmp/reader1-debug | jq -r '.wrap_info.token' > /tmp/reader1-secret

cat /tmp/reader1-debug | jq -r '.data.secret_id' > /tmp/reader1-secret

# check appsecret exists
cat /tmp/reader1-debug
cat /tmp/reader1-secret

curl -v \
   --request POST \
   --header "X-Vault-Token: ${WRITER1_TOKEN}" \
      "${VAULT_ADDRESS}/v1/auth/approle/role/writer1/secret-id" | tee /tmp/writer1-debug

cat /tmp/writer1-debug | jq -r '.data.secret_id' > /tmp/writer1-secret

cat /tmp/writer1-debug
cat /tmp/writer1-secret

VAULT_DOCKER=`docker ps| grep vault | cut -d' ' -f 1`

#echo 'put secret example'
#cat whatever.json | docker exec -i $VAULT_DOCKER vault -v kv put -address=http://127.0.0.1:8200 -mount=secret keys1/zoneName1 raw=-

#echo 'check secret length example'
#docker exec -i $VAULT_DOCKER vault -v kv get -address=http://127.0.0.1:8200 -mount=secret keys1/zoneName1 | wc -l