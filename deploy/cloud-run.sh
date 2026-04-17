#!/usr/bin/env bash
# 一鍵部署到 Google Cloud Run
# 用法：先 export 下方環境變數，然後 ./deploy/cloud-run.sh
#
#   export PROJECT_ID=your-gcp-project
#   export REGION=asia-east1
#   export SERVICE=badminton-bot

set -euo pipefail

: "${PROJECT_ID:?需要設定 PROJECT_ID}"
: "${REGION:=asia-east1}"
: "${SERVICE:=badminton-bot}"

IMAGE="gcr.io/${PROJECT_ID}/${SERVICE}:latest"

echo "▶ Building image ${IMAGE}..."
gcloud builds submit --tag "${IMAGE}" --project "${PROJECT_ID}"

echo "▶ Deploying to Cloud Run..."
gcloud run deploy "${SERVICE}" \
  --image "${IMAGE}" \
  --region "${REGION}" \
  --platform managed \
  --allow-unauthenticated \
  --port 8080 \
  --memory 256Mi \
  --cpu 1 \
  --min-instances 0 \
  --max-instances 2 \
  --set-env-vars "TZ=Asia/Taipei" \
  --set-secrets "LINE_CHANNEL_SECRET=LINE_CHANNEL_SECRET:latest,LINE_CHANNEL_TOKEN=LINE_CHANNEL_TOKEN:latest,LIFF_ID=LIFF_ID:latest,GOOGLE_SHEETS_ID=GOOGLE_SHEETS_ID:latest,TURSO_DATABASE_URL=TURSO_DATABASE_URL:latest,TURSO_AUTH_TOKEN=TURSO_AUTH_TOKEN:latest" \
  --project "${PROJECT_ID}"

echo "✅ Done."
gcloud run services describe "${SERVICE}" --region "${REGION}" --project "${PROJECT_ID}" --format="value(status.url)"
