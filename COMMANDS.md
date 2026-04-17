# 黑武士 Bot 指令速查

> 最後更新：2026-04-18

---

## 📢 公開指令（所有隊員可用）

### `/help` 或 `/說明`
顯示公開指令列表。

**範例：**
```
你：/help
Bot：🏸 羽球成績 Bot 指令：

     /登記 或 /add     → 取得成績登記連結
     /賽事列表         → 顯示目前所有賽事
     /賽事名稱         → 查詢該場比賽所有得獎紀錄（例：/清晨盃）
     /我的ID           → 顯示你的 LINE User ID
     /說明             → 顯示這份說明
```

---

### `/登記` 或 `/add`
取得**一個通用連結**，隊員自己從下拉選賽事。

**範例：**
```
你：/登記
Bot：🏸 成績登記
     👉 https://liff.line.me/2009823919-BC2lq8hx

     從下拉選擇賽事後填寫
```

**使用情境**：隊員往上捲太累找不到 `/開單` 那則，自己打 `/登記` 就拿到連結。

---

### `/所有連結` 或 `/all` 或 `/links`
**列出每場賽事的專屬連結**（比 `/登記` 方便，點連結直接預填賽事）。

**範例：**
```
你：/所有連結
Bot：🏸 目前可登記的賽事：

     ▪️ 2026清晨盃
     👉 https://liff.line.me/2009823919-BC2lq8hx?t=2026清晨盃

     ▪️ 2026會長盃
     👉 https://liff.line.me/2009823919-BC2lq8hx?t=2026會長盃

     ▪️ 2026排名賽
     👉 https://liff.line.me/2009823919-BC2lq8hx?t=2026排名賽
```

**沒有任何賽事時：**
```
你：/所有連結
Bot：（目前沒有賽事）
     管理員用 /新增賽事 或 /開單 新增
```

**使用情境**：現在同時有 3 場賽事在收資料，隊員一次拿所有連結，依自己要的賽事點。

---

### `/賽事列表`
顯示目前下拉選單裡有哪些賽事。

**有資料時：**
```
你：/賽事列表
Bot：📋 目前賽事列表：
     2026清晨盃
     2026會長盃
     2026縣長盃
```

**沒資料時：**
```
你：/賽事列表
Bot：（目前沒有賽事）
     用「/新增賽事 名稱」新增
```

---

### `/<賽事名稱>` — 查詢某場比賽全部得獎紀錄

**範例 1：有紀錄**
```
你：/2026清晨盃
Bot：🏆 2026清晨盃 成績紀錄（5 筆）
     ───────────
     • 香菇｜混雙 冠軍（公開組） / 搭檔:小美
     • 香菇｜男雙 亞軍（甲組） / 搭檔:李振偉
     • 阿強｜男單 四強（公開組）
     • 大頭｜男雙 季軍（乙組） / 搭檔:阿華
     • 小美｜女雙 亞軍（公開組） / 搭檔:阿萍
```

**範例 2：有賽事但沒紀錄**
```
你：/2026會長盃
Bot：📋 2026會長盃
     （目前沒有紀錄）
```

**範例 3：打亂碼（不在賽事列表內）**
```
你：/2027地獄盃
Bot：（靜默，bot 不回）
```

> 💡 只有「已註冊的賽事名稱」才會回應。避免隊員 typo 洗版。

**範例 4：紀錄超過 40 筆**
```
你：/2026清晨盃
Bot：🏆 2026清晨盃 成績紀錄（67 筆）
     ───────────
     • 香菇｜混雙 冠軍（公開組） / 搭檔:小美
     ...（40 筆後截斷）...
     ...（另有 27 筆，請至試算表檢視）
```

---

### `/我的ID` 或 `/myid`
顯示自己的 LINE User ID（想申請當管理員時用）。

**正常：**
```
你：/我的ID
Bot：你的 LINE User ID：
     U1a2b3c4d5e6f7890abcdef1234567890
```

**異常（在群組內但沒加 bot 好友）：**
```
你：/我的ID
Bot：（拿不到你的 ID，請先加 bot 為好友）
```

---

## 🔐 管理員指令

### `/dhelp`
顯示管理員專屬指令列表。

**你是管理員：**
```
你：/dhelp
Bot：🔐 管理員指令：

     /開單 賽事名稱     → 發起成績登記（會自動加入列表）
     /新增賽事 XXX      → 新增賽事到表單下拉
     /刪除賽事 XXX      → 從表單下拉移除
     /新增管理員 Uxxx   → 新增管理員
     /刪除管理員 Uxxx   → 移除管理員
     /管理員列表        → 顯示目前所有管理員
```

**你不是管理員：**
```
你：/dhelp
Bot：（靜默，不洩漏有這個指令）
```

---

### `/開單 <賽事名稱>`

**正常：**
```
你：/開單 2026清晨盃
Bot：📋 2026清晨盃 成績登記
     👉 https://liff.line.me/2009823919-BC2lq8hx?t=2026清晨盃

     （每人可填多項，可隨時回來修改）
```
> 如果 `2026清晨盃` 還不在下拉裡，會自動加進去。

**沒帶賽事名稱：**
```
你：/開單
Bot：用法：/開單 2026清晨杯
```

**非管理員：**
```
隊員：/開單 2026清晨盃
Bot：⚠️ 只有管理員可以開單
```

---

### `/新增賽事 <名稱>`
加到下拉清單（不發連結）。

```
你：/新增賽事 2026會長盃
Bot：✅ 已新增賽事：
     2026會長盃
```

**賽事已存在：**
```
你：/新增賽事 2026清晨盃
Bot：✅ 已新增賽事：
     2026清晨盃
```
> 重複新增不報錯，就當成功。

---

### `/刪除賽事 <名稱>`
從下拉清單移除。**已填的舊成績不會被刪**。

```
你：/刪除賽事 2026會長盃
Bot：🗑 已刪除賽事：
     2026會長盃
```

**使用情境**：某個賽事打完一年後，下拉選單太雜，把過期的移掉。

---

### `/新增管理員 <UserID>`

**流程：**
1. 對方加 bot 為好友
2. 對方傳 `/我的ID` → 拿到 `U1a2b3c4...`
3. 對方把 ID 私訊給你
4. 你打 `/新增管理員 U1a2b3c4...`

**範例：**
```
你：/新增管理員 U9999abcdef1234567890
Bot：✅ 已新增管理員：
     U9999abcdef1234567890
```

**沒帶 ID：**
```
你：/新增管理員
Bot：用法：/新增管理員 Uxxxxxxxx（該使用者的 LINE User ID）
```

---

### `/刪除管理員 <UserID>`

```
你：/刪除管理員 U9999abcdef1234567890
Bot：🗑 已刪除管理員：
     U9999abcdef1234567890
```

> ⚠️ **別把自己刪了**。如果真的刪了，去 Google Sheet 的 `admins` 分頁把 UserID 那列清空，再打一次 `/addme` 重新 bootstrap。

---

### `/管理員列表`

```
你：/管理員列表
Bot：👮 管理員列表：
     • 蔡秉諺（U1a2b3c4...）
     • 林大華（U9999abcd...）
```

**Store 還沒有管理員但你是 env 設的靜態管理員：**
```
你：/管理員列表
Bot：（Store 內沒有管理員；只靠環境變數 ADMIN_USER_IDS）
```

---

## 🔑 隱藏指令（不列在任何 help）

### `/addme`
**第一位管理員自助 bootstrap 專用**。

**第一次（admins 是空的）：**
```
你：/addme
Bot：✅ 你已成為管理員：蔡秉諺

     輸入 /dhelp 查看管理員指令
```

**之後（已有管理員存在）：**
```
你：/addme
Bot：（靜默，不回）
```

**在群組打但沒加 bot 好友（拿不到 userID）：**
```
你：/addme
Bot：⚠️ 請先加 bot 為好友，再試一次
```

> 🔐 **安全設計**：只有 Store 裡的 admins 是空的時候才生效。確保只有「第一個想當管理員」的人能用這招。之後要加管理員只能透過現有管理員下指令。

---

## 🎯 完整使用情境

### 情境 1：首次部署，設第一位管理員

```
Step 1：你在 LINE 搜尋 @977pgiqm 或掃 QR → 加 bot 好友

Step 2：
你：/addme
Bot：✅ 你已成為管理員：蔡秉諺
     輸入 /dhelp 查看管理員指令

Step 3（可選）：驗證
你：/管理員列表
Bot：👮 管理員列表：
     • 蔡秉諺（U1234...）
```

---

### 情境 2：週末比完清晨盃，開始蒐集成績

```
Step 1：管理員群組開單
你（管理員）：/開單 2026清晨盃
Bot：📋 2026清晨盃 成績登記
     👉 https://liff.line.me/...

Step 2：隊員 A 填寫
隊員 A：[點連結] → 填表 → 送出：
  姓名：香菇
  賽事：2026清晨盃
  組別：公開
  項目：混雙
  搭檔：小美
  名次：冠軍
→ ✅ 已送出
→ 按「再新增一筆」
→ 姓名「香菇」賽事「2026清晨盃」組別「公開」保留
→ 填：男雙 亞軍 搭檔李振偉
→ ✅ 已送出

Step 3：其他隊員陸續填

Step 4：結算（隔天）
你：/2026清晨盃
Bot：🏆 2026清晨盃 成績紀錄（12 筆）...
```

---

### 情境 3：新隊員加入想當共同管理員

```
Step 1：新管理員先加 bot 好友

Step 2：新管理員取得自己 ID
新人 → bot：/我的ID
Bot → 新人：你的 LINE User ID：
            U9999abcdef...

Step 3：新人把 ID 私訊給你（Copy 貼上）

Step 4：你加權限
你：/新增管理員 U9999abcdef...
Bot：✅ 已新增管理員：U9999abcdef...

Step 5：新人驗證
新人：/dhelp
Bot：🔐 管理員指令：...
```

---

### 情境 4：隊員不小心漏掉 /開單 訊息

```
隊員：/登記
Bot：🏸 成績登記
     👉 https://liff.line.me/...

隊員：[點連結] → 從下拉選「2026清晨盃」→ 填表 → 送出
```

---

### 情境 5：比完賽後要統計

```
Step 1：快速看得獎清單
你：/2026清晨盃
Bot：🏆 2026清晨盃 成績紀錄（12 筆）
     • 香菇｜混雙 冠軍（公開組）
     ...

Step 2：要做年度統計 → 開 Google Sheet
→ 插入 → 樞紐分析表
→ 列：賽事 + 項目
→ 欄：名次
→ 值：姓名（計數）

就會看到類似：

                       冠軍  亞軍  季軍  四強  八強
2026清晨盃
  混雙                  2     1     0     1     0
  男雙                  1     0     1     0     0
2026會長盃
  ...
```

---

## 🗃️ Google Sheet 分頁說明

Bot 會自動在你的 Sheet 建立：

| 分頁 | 欄位 | 用途 | 備註 |
|------|-----|------|------|
| `2026`（當年度）| 時間 / 輸入人 / 姓名 / 賽事 / 組別 / 項目 / 搭檔 / 名次 / 備註 | 成績資料 | 每年自動新建 `2027`、`2028`... |
| `tournaments` | 賽事 | 賽事下拉清單 | 可手動在 Sheet 編輯 |
| `admins` | UserID / Name / AddedAt | 管理員清單 | 緊急時可手動清空 |

---

## 🚨 緊急情況 SOP

### 「我把自己從管理員刪了！」
1. 開 Google Sheet → `admins` 分頁
2. 把整張表的內容清空（只留 header）
3. 回 bot，傳 `/addme` → 你又是唯一管理員了

### 「Bot 沒反應」
1. 檢查 Cloud Run 狀態：https://console.cloud.google.com/run/detail/asia-east1/badminton-bot
2. 看 log：`gcloud run services logs read badminton-bot --region asia-east1 --limit 50`
3. 如果 deploy 失敗，看 Cloud Build：https://console.cloud.google.com/cloud-build/builds

### 「隊員填了但 Sheet 沒資料」
- 檢查 Service Account 是否還是 `badminton-bot-sa@badminton-bot-493614.iam.gserviceaccount.com`
- 確認那個 SA 還是 Sheet 的編輯者

### 「金鑰外洩」
- LINE token → Developers console → Messaging API → Reissue
- Turso token → Turso dashboard → Revoke + Create new
- 都要同步更新 Secret Manager 的對應 secret

---

## 🛠️ 技術備註

| 項目 | 值 |
|------|-----|
| LINE Channel | `@977pgiqm`（黑武士bot） |
| LIFF ID | `2009823919-BC2lq8hx` |
| GCP 專案 | `badminton-bot-493614` |
| GCP 專案編號 | `320072287872` |
| Service URL | https://badminton-bot-320072287872.asia-east1.run.app |
| Webhook URL | `{service}/webhook` |
| LIFF Endpoint | `{service}/liff/index.html` |
| Region | `asia-east1`（台灣） |
| GitHub Repo | https://github.com/aqws11223344/dark-badmintonteam |
| CI/CD | push main 自動部署（Cloud Build） |
| Sheet ID | `1FRexQNQAicpDlsTqkE17OBCHQVQ9ZVf6SdI495si0D8` |

---

## 🔧 部署後要設的環境變數（Cloud Run 主控台）

| 變數 | 說明 | 是否必填 |
|------|-----|---------|
| `LINE_CHANNEL_SECRET` | LINE channel secret | ✅ |
| `LINE_CHANNEL_TOKEN` | LINE access token | ✅ |
| `LIFF_ID` | LIFF app ID | ✅ |
| `GOOGLE_SHEETS_ID` | Sheet ID | ✅ |
| `GOOGLE_APPLICATION_CREDENTIALS` | SA JSON 路徑（`/secrets/sa.json`） | ✅ |
| `TURSO_DATABASE_URL` | Turso DB URL | 可選（目前 MVP 未啟用） |
| `TURSO_AUTH_TOKEN` | Turso token | 可選 |
| `ADMIN_USER_IDS` | 靜態管理員 ID（逗號分隔） | 可選（也可用 /addme） |
| `BOOTSTRAP_TOKEN` | 自訂 bootstrap token | 可選（目前寫死 /addme） |
