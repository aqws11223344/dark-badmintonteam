#!/usr/bin/env bash
# 一鍵在 Cloud Shell 部署整個專案。
# 前提：~/.env 和 ~/service-account.json 已上傳好，且此 repo 已 clone 到 ~/dark-badmintonteam
#
# 用法：
#   cd ~/dark-badmintonteam
#   bash deploy/setup-and-deploy.sh
#
# 可以重複跑（idempotent）。

set -euo pipefail

PROJECT_ID="${PROJECT_ID:-badminton-bot-493614}"
REGION="${REGION:-asia-east1}"
SERVICE="${SERVICE:-badminton-bot}"

echo "▶ 使用 GCP 專案: $PROJECT_ID"
gcloud config set project "$PROJECT_ID"

echo "▶ 搬移 ~/.env 和 ~/service-account.json 進專案（若在家目錄）"
[ -f ~/.env ] && mv -n ~/.env . || true
[ -f ~/service-account.json ] && mv -n ~/service-account.json . || true

if [ ! -f .env ] || [ ! -f service-account.json ]; then
    echo "❌ 找不到 .env 或 service-account.json，請確認已上傳到 ~/ 或放在專案根目錄"
    exit 1
fi

echo "▶ 啟用必要的 GCP API"
gcloud services enable \
    run.googleapis.com \
    cloudbuild.googleapis.com \
    secretmanager.googleapis.com \
    artifactregistry.googleapis.com \
    sheets.googleapis.com \
    drive.googleapis.com \
    --quiet

echo "▶ 讀取 .env"
set -a
# shellcheck disable=SC1091
source .env
set +a

upsert_secret() {
    local name="$1" value="$2"
    if gcloud secrets describe "$name" >/dev/null 2>&1; then
        printf '%s' "$value" | gcloud secrets versions add "$name" --data-file=- --quiet
    else
        printf '%s' "$value" | gcloud secrets create "$name" --data-file=- --quiet
    fi
}

upsert_secret_file() {
    local name="$1" file="$2"
    if gcloud secrets describe "$name" >/dev/null 2>&1; then
        gcloud secrets versions add "$name" --data-file="$file" --quiet
    else
        gcloud secrets create "$name" --data-file="$file" --quiet
    fi
}

echo "▶ 上傳 secrets 到 Secret Manager"
upsert_secret LINE_CHANNEL_SECRET "$LINE_CHANNEL_SECRET"
upsert_secret LINE_CHANNEL_TOKEN "$LINE_CHANNEL_TOKEN"
upsert_secret LIFF_ID "$LIFF_ID"
upsert_secret GOOGLE_SHEETS_ID "$GOOGLE_SHEETS_ID"
upsert_secret TURSO_DATABASE_URL "$TURSO_DATABASE_URL"
upsert_secret TURSO_AUTH_TOKEN "$TURSO_AUTH_TOKEN"
upsert_secret_file SA_JSON service-account.json

echo "▶ 授權 Cloud Run 預設 SA 讀取 secrets"
PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format="value(projectNumber)")
COMPUTE_SA="${PROJECT_NUMBER}-compute@developer.gserviceaccount.com"

for S in LINE_CHANNEL_SECRET LINE_CHANNEL_TOKEN LIFF_ID GOOGLE_SHEETS_ID TURSO_DATABASE_URL TURSO_AUTH_TOKEN SA_JSON; do
    gcloud secrets add-iam-policy-binding "$S" \
        --member="serviceAccount:${COMPUTE_SA}" \
        --role="roles/secretmanager.secretAccessor" \
        --quiet >/dev/null 2>&1 || true
done

echo "▶ 部署到 Cloud Run（第一次約 3-5 分鐘）"
gcloud run deploy "$SERVICE" \
    --source . \
    --region "$REGION" \
    --platform managed \
    --allow-unauthenticated \
    --port 8080 \
    --memory 256Mi \
    --cpu 1 \
    --min-instances 0 \
    --max-instances 2 \
    --timeout 30 \
    --set-env-vars "TZ=Asia/Taipei,GOOGLE_APPLICATION_CREDENTIALS=/secrets/sa.json" \
    --set-secrets "LINE_CHANNEL_SECRET=LINE_CHANNEL_SECRET:latest,LINE_CHANNEL_TOKEN=LINE_CHANNEL_TOKEN:latest,LIFF_ID=LIFF_ID:latest,GOOGLE_SHEETS_ID=GOOGLE_SHEETS_ID:latest,TURSO_DATABASE_URL=TURSO_DATABASE_URL:latest,TURSO_AUTH_TOKEN=TURSO_AUTH_TOKEN:latest,/secrets/sa.json=SA_JSON:latest" \
    --quiet

# 優先使用新格式 URL（{service}-{project_number}.{region}.run.app），後備 status.url
URL="https://${SERVICE}-${PROJECT_NUMBER}.${REGION}.run.app"
if ! curl -sf -o /dev/null -w "%{http_code}" "$URL" 2>/dev/null | grep -q "^[23]"; then
    # 新 URL 拿不到 2xx/3xx 就退回 status.url
    FALLBACK=$(gcloud run services describe "$SERVICE" --region "$REGION" --format="value(status.url)")
    [ -n "$FALLBACK" ] && URL="$FALLBACK"
fi

echo ""
echo "=================================================="
echo "✅ 部署完成！"
echo "=================================================="
echo ""
echo "服務網址: $URL"
echo ""
echo "請到以下兩個地方更新 URL："
echo ""
echo "1. LINE Developers → 黑武士bot → Messaging API → Webhook URL："
echo "   $URL/webhook"
echo "   （並記得把 'Use webhook' 打開）"
echo ""
echo "2. LINE Developers → 比賽成績登記(Login channel) → LIFF → Endpoint URL："
echo "   $URL/liff/index.html"
echo ""
echo "=================================================="
