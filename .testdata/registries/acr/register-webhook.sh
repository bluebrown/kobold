#!/usr/bin/env bash

set -euo pipefail

az acr webhook create -n "$ACR_WEBHOOK_NAME" -r "$ACR_NAME" \
  --uri "${ACR_WEBHOOK_HOST}${ACR_WEBHOOK_ENDPOINT}" --actions push \
  --headers "Authorization=Kobold $ACR_TOKEN"
