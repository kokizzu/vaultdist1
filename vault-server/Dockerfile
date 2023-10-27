
FROM hashicorp/vault

RUN apk add --no-cache bash jq

COPY plugins /plugins
COPY reader1-policy.hcl /vault/config/reader1-policy.hcl
COPY writer1-policy.hcl /vault/config/writer1-policy.hcl
COPY init_vault.sh /init_vault.sh

EXPOSE 8200

ENTRYPOINT [ "/init_vault.sh" ]

HEALTHCHECK \
    --start-period=5s \
    --interval=1s \
    --timeout=1s \
    --retries=30 \
        CMD [ "/bin/sh", "-c", "[ -f /tmp/healthy ]" ]