# 🏸 Badminton Bot

LINE Bot + LIFF 表單，蒐集羽球隊隊員的賽事成績，雙寫到 Google Sheets（給教練看）和 Turso SQLite（給 bot 查詢）。

## 架構

```
LINE 群組
   │ 隊長：/開單 2026清晨杯
   ▼
LINE 平台 ──webhook──▶ Go 伺服器（Cloud Run）
                         │
                         ├─ 回傳訊息：📋 點我填寫 [LIFF 連結]
                         │
                         ▼
                       隊員點 → LIFF 表單
                         │ POST /api/results
                         ▼
                       Turso (主)  ──雙寫──▶ Google Sheets (教練看)
```

## 專案結構

```
badminton-bot/
├── cmd/server/main.go          # 入口、Gin 路由
├── internal/
│   ├── config/                 # 環境變數載入
│   ├── domain/                 # MatchResult、Options
│   ├── line/                   # webhook 處理、LIFF API
│   └── store/
│       ├── store.go            # 介面 + in-memory fallback
│       ├── sheets/             # Google Sheets 實作
│       ├── turso/              # libsql 實作
│       └── dual/               # 雙寫 wrapper
├── web/liff/index.html         # LIFF 表單頁
├── deploy/cloud-run.sh         # 一鍵部署
├── Dockerfile
└── .env.example
```

---

## 1. 需要先準備什麼

> 全部都做完 README 後面的步驟就可以填到 `.env`

### A. LINE Developers
1. 到 https://developers.line.biz/console 建立 **Provider**
2. 建立 **Messaging API channel**
   - 拿到：`Channel secret`、`Channel access token (long-lived)`
3. 在同一個 channel 底下建立 **LIFF app**
   - Endpoint URL：`https://你的-cloud-run-網址/liff/index.html`
   - Size：`Full`
   - Scope：`profile`、`openid`
   - 拿到：`LIFF ID`（格式像 `1234567890-AbCdEfGh`）
4. 在 channel 設定：
   - 關閉「自動回覆訊息」
   - 開啟「Webhook」
   - Webhook URL：`https://你的-cloud-run-網址/webhook`

### B. Google Sheets（給教練看的鏡像）
1. 開一份新的 Google Sheet，把第一個分頁改名為 `results`
2. 從網址抓 Sheet ID：`docs.google.com/spreadsheets/d/{SHEET_ID}/edit`
3. 到 https://console.cloud.google.com 建立 service account：
   - 啟用 **Google Sheets API**
   - 建立 service account → 下載 JSON 金鑰存成 `service-account.json`
4. **把 service account email 加到那份 Sheet 的共用權限（編輯者）**

### C. Turso（bot 自己查詢用）
1. 到 https://turso.tech 註冊（GitHub 登入）
2. `turso db create badminton`
3. `turso db show badminton --url` → `TURSO_DATABASE_URL`
4. `turso db tokens create badminton` → `TURSO_AUTH_TOKEN`

### D. Google Cloud（部署用）
1. 建立 GCP 專案，啟用：
   - Cloud Run API
   - Cloud Build API
   - Secret Manager API
2. 安裝 [gcloud CLI](https://cloud.google.com/sdk/docs/install)
3. `gcloud auth login` 並 `gcloud config set project YOUR_PROJECT_ID`

---

## 2. 本機跑起來

```bash
cd C:\ownproject\badminton-bot

# 改 module 名稱（把 yourname 換成你的 GitHub 帳號或任意名稱）
# 編輯 go.mod 第一行 + 全部 go 檔的 import 路徑

cp .env.example .env
# 編輯 .env 填入上面拿到的所有金鑰

go mod tidy
go run ./cmd/server
```

伺服器會跑在 http://localhost:8080。

### 用 ngrok 把 webhook 暴露到公網（測試用）
```bash
ngrok http 8080
# 把 https://xxxx.ngrok.io/webhook 設到 LINE channel 的 Webhook URL
```

---

## 3. 部署到 Cloud Run

### 把所有 secrets 上傳到 Secret Manager
```bash
# 一個一個建立
gcloud secrets create LINE_CHANNEL_SECRET --replication-policy=automatic
echo -n "你的 channel secret" | gcloud secrets versions add LINE_CHANNEL_SECRET --data-file=-

gcloud secrets create LINE_CHANNEL_TOKEN --replication-policy=automatic
echo -n "你的 channel token" | gcloud secrets versions add LINE_CHANNEL_TOKEN --data-file=-

# 重複給 LIFF_ID, GOOGLE_SHEETS_ID, TURSO_DATABASE_URL, TURSO_AUTH_TOKEN
```

### 給 Cloud Run service account 讀取 secret 的權限
```bash
PROJECT_NUMBER=$(gcloud projects describe YOUR_PROJECT_ID --format="value(projectNumber)")
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:${PROJECT_NUMBER}-compute@developer.gserviceaccount.com" \
  --role="roles/secretmanager.secretAccessor"
```

### Sheets 認證（兩種選一個）
- **方案 A（簡單）**：把 `service-account.json` 內容也丟進 Secret Manager，部署時掛載成檔案
- **方案 B（推薦）**：把那個 service account 直接設為 Cloud Run 的執行身分，並把 Sheet 共用權限給它，這樣不用 JSON 檔

### 一鍵部署
```bash
export PROJECT_ID=your-gcp-project
export REGION=asia-east1
chmod +x deploy/cloud-run.sh
./deploy/cloud-run.sh
```

部署完會印出網址，把它填回：
- LINE channel 的 **Webhook URL**：`https://網址/webhook`
- LIFF 的 **Endpoint URL**：`https://網址/liff/index.html`

---

## 4. 使用方式

把 bot 加到球隊群組，然後：

| 指令 | 效果 |
|------|------|
| `/開單 2026清晨杯` | 群組裡發出一則含 LIFF 連結的訊息 |
| `/說明` | 顯示用法 |

隊員點連結 → 在 LINE 內開啟表單 → 填項目（可連填多筆）→ 送出。

---

## 5. 之後可以擴充的方向

- `/結算 2026清晨杯` → bot 整理成獎牌榜貼回群組
- `/我的成績` → 個人聊天回傳獎牌牆
- 球隊排行榜（年度金牌數）
- 自動提醒沒填的人
- 把選項管理介面也做成網頁（目前是寫死在 `internal/domain/result.go`）

---

## 6. 小備註

- Module 名稱目前是 `github.com/aqws11223344/dark-badmintonteam`，請改成你自己的（影響所有 import）
- LINE SDK v8 的 webhook 套件 API 偶爾會微調，若 build 失敗看 godoc 對一下型別
- Cloud Run 預設會在沒流量時 scale to 0，第一個請求會冷啟（約 1-3 秒），LINE webhook timeout 是 10 秒，沒問題
