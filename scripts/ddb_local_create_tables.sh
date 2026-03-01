#!/usr/bin/env bash
set -euo pipefail

: "${DYNAMODB_ENDPOINT:?set DYNAMODB_ENDPOINT}"
: "${AWS_REGION:?set AWS_REGION}"

create_pk_only () {
  local name="$1"
  aws dynamodb create-table \
    --endpoint-url "$DYNAMODB_ENDPOINT" \
    --region "$AWS_REGION" \
    --table-name "$name" \
    --attribute-definitions AttributeName=pk,AttributeType=S \
    --key-schema AttributeName=pk,KeyType=HASH \
    --billing-mode PAY_PER_REQUEST \
    >/dev/null
}

create_pk_sk () {
  local name="$1"
  aws dynamodb create-table \
    --endpoint-url "$DYNAMODB_ENDPOINT" \
    --region "$AWS_REGION" \
    --table-name "$name" \
    --attribute-definitions AttributeName=pk,AttributeType=S AttributeName=sk,AttributeType=S \
    --key-schema AttributeName=pk,KeyType=HASH AttributeName=sk,KeyType=RANGE \
    --billing-mode PAY_PER_REQUEST \
    >/dev/null
}

create_pk_only "${DDB_ACCOUNTS_TABLE:-accounts}" || true
create_pk_only "${DDB_IDENTITIES_TABLE:-auth_identities}" || true
create_pk_only "${DDB_SESSIONS_TABLE:-sessions}" || true
create_pk_sk   "${DDB_AUDIT_EVENTS_TABLE:-audit_events}" || true

create_pk_only "${DDB_HOSTING_UPLOADS_TABLE:-hosting_uploads}" || true
create_pk_only "${DDB_HOSTING_SITES_TABLE:-hosting_sites}" || true
create_pk_only "${DDB_HOSTING_RELEASES_TABLE:-hosting_releases}" || true

echo "done"
