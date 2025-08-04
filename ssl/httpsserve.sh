#!/usr/bin/env bash

set -euo pipefail

#--------------- configuration ------------------------------------------------
SUBDOMAIN="${1:-anonymoususer}"
FULLDOMAIN="${SUBDOMAIN}.nexasdk.com"
API="https://api.cloudflare.com/client/v4"
HEADER_AUTH="Authorization: Bearer ${CF_Token:?need CF_Token}"
HEADER_JSON="Content-Type: application/json"
TTL=60

#--------------- locate / install acme.sh ------------------------------------
ACME="$HOME/acme.sh/acme.sh"
if [[ ! -x "$ACME" ]]; then
  echo "acme.sh not found – installing…" >&2
  curl -fsSL https://get.acme.sh | sh >/dev/null 2>&1
  ACME="$HOME/.acme.sh/acme.sh"
  echo "acme.sh installed to $ACME" >&2
fi

# Abort if still not executable
[[ -x "$ACME" ]] || { echo "Failed to install acme.sh" >&2; exit 1; }

#--------------- determine private IP ----------------------------------------
PRIVATE_IP=$(ipconfig getifaddr en0 2>/dev/null || ipconfig getifaddr en1)
if [[ -z "$PRIVATE_IP" ]]; then
  echo "Could not determine private IP address" >&2
  exit 1
fi

#--------------- cleanup ------------------------------------------------------
cleanup() {
  lookup=$(curl -sS "$API/zones/$CF_Zone_ID/dns_records?type=A&name=$FULLDOMAIN" -H "$HEADER_AUTH")
  rid=$(jq -r '.result[0].id // empty' <<<"$lookup")

  if [[ -z "$rid" ]]; then
    echo "$FULLDOMAIN is already clean."
  else
    curl -sS -X DELETE "$API/zones/$CF_Zone_ID/dns_records/$rid" -H "$HEADER_AUTH" >/dev/null \
      && echo "$FULLDOMAIN is cleaned up."
  fi
}
cleanup
trap cleanup EXIT

#--------------- create DNS record -------------------------------------------
printf "Creating DNS record %s → %s\n" "$FULLDOMAIN" "$PRIVATE_IP"
create_resp=$(
  curl -sS -X POST "$API/zones/$CF_Zone_ID/dns_records" \
       -H "$HEADER_AUTH" -H "$HEADER_JSON" \
       --data '{"type":"A","name":"'$SUBDOMAIN'","content":"'$PRIVATE_IP'","ttl":'$TTL',"proxied":false}'
)
record_id=$(jq -r '.result.id // empty' <<<"$create_resp")
if [[ -z "$record_id" ]]; then
  echo "Failed to create record:" >&2
  jq -r '.errors[]?.message' <<<"$create_resp" >&2 || echo "$create_resp" >&2
  exit 1
fi
printf "Record created (id=%s)\n" "$record_id"s


#--------------- generate certificate ----------------------------------------
# Ensure ssl directory exists
mkdir -p ssl
"$ACME" --server letsencrypt --issue --dns dns_cf -d "$FULLDOMAIN" --force
"$ACME" --install-cert -d "$FULLDOMAIN" --key-file ssl/key.pem --cert-file ssl/cert.pem --force

#--------------- run server ---------------------------------------------------
echo "Serving at: https://$FULLDOMAIN:18181/docs/ui"
./build/nexa serve --host=0.0.0.0:18181 --https --certfile ssl/cert.pem --keyfile ssl/key.pem