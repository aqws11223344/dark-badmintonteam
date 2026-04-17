// Package line 處理 LINE webhook 事件，並提供 LIFF 表單後端 API。
package line

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
	"github.com/line/line-bot-sdk-go/v8/linebot/webhook"

	"github.com/aqws11223344/dark-badmintonteam/internal/domain"
	"github.com/aqws11223344/dark-badmintonteam/internal/store"
)

type Config struct {
	ChannelSecret  string
	ChannelToken   string
	LIFFID         string
	SheetsID       string   // Google Sheet ID（/表單 指令用）
	AdminUserIDs   []string // 靜態清單（env var），永遠是 admin
	BootstrapToken string   // 秘密指令，第一位 admin 用
	Store          store.Store
}

type Handler struct {
	cfg     Config
	bot     *messaging_api.MessagingApiAPI
	staticAdmins map[string]bool
}

func New(cfg Config) (*Handler, error) {
	bot, err := messaging_api.NewMessagingApiAPI(cfg.ChannelToken)
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(cfg.AdminUserIDs))
	for _, id := range cfg.AdminUserIDs {
		m[id] = true
	}
	return &Handler{cfg: cfg, bot: bot, staticAdmins: m}, nil
}

// isAdmin 先查靜態清單，再查 Store 動態清單。
func (h *Handler) isAdmin(ctx context.Context, userID string) bool {
	if userID == "" {
		return false
	}
	if h.staticAdmins[userID] {
		return true
	}
	admins, err := h.cfg.Store.ListAdmins(ctx)
	if err != nil {
		log.Printf("list admins failed: %v", err)
		return false
	}
	for _, a := range admins {
		if a.UserID == userID {
			return true
		}
	}
	return false
}

// ServeHTTP 處理 LINE 的 webhook callback（含簽章驗證）。
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cb, err := webhook.ParseRequest(h.cfg.ChannelSecret, r)
	if err != nil {
		log.Printf("parse webhook: %v", err)
		if errors.Is(err, webhook.ErrInvalidSignature) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	for _, ev := range cb.Events {
		switch e := ev.(type) {
		case webhook.MessageEvent:
			if msg, ok := e.Message.(webhook.TextMessageContent); ok {
				h.handleText(e.ReplyToken, msg.Text, sourceUserID(e.Source))
			}
		case webhook.FollowEvent:
			h.reply(e.ReplyToken, "歡迎加入！\n在群組裡輸入「/開單 賽事名稱」開始蒐集成績 🏸")
		}
	}
	w.WriteHeader(http.StatusOK)
}

// sourceUserID 從事件來源抽出 LINE userID（可能是 1 對 1 聊天、群組、多人聊天室）。
// 若使用者沒加 bot 為好友，group/room source 的 userID 可能為空。
func sourceUserID(src webhook.SourceInterface) string {
	switch s := src.(type) {
	case webhook.UserSource:
		return s.UserId
	case webhook.GroupSource:
		return s.UserId
	case webhook.RoomSource:
		return s.UserId
	}
	return ""
}

// stripAnyPrefix 若 text 開頭是任一 prefix，回傳去掉後的 trim 內容與 true。
func stripAnyPrefix(text string, prefixes ...string) (string, bool) {
	for _, p := range prefixes {
		if strings.HasPrefix(text, p) {
			return strings.TrimSpace(strings.TrimPrefix(text, p)), true
		}
	}
	return "", false
}

// matchesAny 檢查 text 是否等於任一候選字串。
func matchesAny(text string, candidates ...string) bool {
	for _, c := range candidates {
		if text == c {
			return true
		}
	}
	return false
}

func (h *Handler) handleText(replyToken, text, userID string) {
	text = strings.TrimSpace(text)
	ctx := context.Background()

	switch {
	// ===== 秘密 bootstrap：/addme（不列在 help）=====
	// 只有當 Store 裡還沒任何 admin 時才有效。
	case text == "/addme":
		h.handleBootstrap(ctx, replyToken, userID)

	case matchesAny(text, "/我的ID", "/myid", "/id"):
		if userID == "" {
			h.reply(replyToken, "（拿不到你的 ID，請先加 bot 為好友）")
			return
		}
		h.reply(replyToken, "你的 LINE User ID：\n"+userID)

	case matchesAny(text, "/登記", "/add", "/register"):
		if h.cfg.LIFFID == "" {
			h.reply(replyToken, "尚未設定 LIFF_ID")
			return
		}
		url := fmt.Sprintf("https://liff.line.me/%s", h.cfg.LIFFID)
		h.reply(replyToken, "🏸 成績登記\n👉 "+url+"\n\n從下拉選擇賽事後填寫")

	case matchesAny(text, "/所有連結", "/links", "/all"):
		if h.cfg.LIFFID == "" {
			h.reply(replyToken, "尚未設定 LIFF_ID")
			return
		}
		tournaments, err := h.cfg.Store.ListTournaments(ctx)
		if err != nil {
			h.reply(replyToken, "❌ 查詢失敗："+err.Error())
			return
		}
		if len(tournaments) == 0 {
			h.reply(replyToken, "（目前沒有賽事）\n管理員用 /新增賽事 或 /開單 新增")
			return
		}
		lines := []string{"🏸 目前可登記的賽事："}
		for _, t := range tournaments {
			lines = append(lines, "")
			lines = append(lines, "▪️ "+t)
			lines = append(lines, "👉 https://liff.line.me/"+h.cfg.LIFFID+"?t="+t)
		}
		h.reply(replyToken, strings.Join(lines, "\n"))

	case strings.HasPrefix(text, "/開單") || strings.HasPrefix(text, "/open"):
		name, _ := stripAnyPrefix(text, "/開單", "/open")
		if name == "" {
			h.reply(replyToken, "用法：/開單 2026清晨杯（或 /open 2026清晨杯）")
			return
		}
		if !h.isAdmin(ctx, userID) {
			h.reply(replyToken, "⚠️ 只有管理員可以開單")
			return
		}
		if err := h.cfg.Store.AddTournament(ctx, name); err != nil {
			log.Printf("add tournament on /開單 failed: %v", err)
		}
		h.replyWithLIFF(replyToken, name)

	case strings.HasPrefix(text, "/新增賽事") || strings.HasPrefix(text, "/addt"):
		name, _ := stripAnyPrefix(text, "/新增賽事", "/addt")
		if name == "" {
			h.reply(replyToken, "用法：/新增賽事 2026清晨杯（或 /addt 2026清晨杯）")
			return
		}
		if !h.isAdmin(ctx, userID) {
			h.reply(replyToken, "⚠️ 只有管理員可以新增賽事")
			return
		}
		if err := h.cfg.Store.AddTournament(ctx, name); err != nil {
			h.reply(replyToken, "❌ 新增失敗："+err.Error())
			return
		}
		h.reply(replyToken, "✅ 已新增賽事：\n"+name)

	case strings.HasPrefix(text, "/刪除賽事") || strings.HasPrefix(text, "/delt"):
		name, _ := stripAnyPrefix(text, "/刪除賽事", "/delt")
		if name == "" {
			h.reply(replyToken, "用法：/刪除賽事 2026清晨杯（或 /delt 2026清晨杯）")
			return
		}
		if !h.isAdmin(ctx, userID) {
			h.reply(replyToken, "⚠️ 只有管理員可以刪除賽事")
			return
		}
		if err := h.cfg.Store.RemoveTournament(ctx, name); err != nil {
			h.reply(replyToken, "❌ 刪除失敗："+err.Error())
			return
		}
		h.reply(replyToken, "🗑 已刪除賽事：\n"+name)

	// ===== 管理員管理 =====
	case strings.HasPrefix(text, "/新增管理員") || strings.HasPrefix(text, "/addadmin"):
		target, _ := stripAnyPrefix(text, "/新增管理員", "/addadmin")
		if target == "" {
			h.reply(replyToken, "用法：/新增管理員 Uxxxxxxxx（或 /addadmin Uxxxxxxxx）")
			return
		}
		if !h.isAdmin(ctx, userID) {
			h.reply(replyToken, "⚠️ 只有管理員可以新增管理員")
			return
		}
		if err := h.cfg.Store.AddAdmin(ctx, store.Admin{UserID: target, AddedAt: time.Now()}); err != nil {
			h.reply(replyToken, "❌ 新增失敗："+err.Error())
			return
		}
		h.reply(replyToken, "✅ 已新增管理員：\n"+target)

	case strings.HasPrefix(text, "/刪除管理員") || strings.HasPrefix(text, "/deladmin"):
		target, _ := stripAnyPrefix(text, "/刪除管理員", "/deladmin")
		if target == "" {
			h.reply(replyToken, "用法：/刪除管理員 Uxxxxxxxx（或 /deladmin Uxxxxxxxx）")
			return
		}
		if !h.isAdmin(ctx, userID) {
			h.reply(replyToken, "⚠️ 只有管理員可以刪除管理員")
			return
		}
		if err := h.cfg.Store.RemoveAdmin(ctx, target); err != nil {
			h.reply(replyToken, "❌ 刪除失敗："+err.Error())
			return
		}
		h.reply(replyToken, "🗑 已刪除管理員：\n"+target)

	case text == "/表單" || text == "/sheet":
		if !h.isAdmin(ctx, userID) {
			return // 靜默
		}
		if h.cfg.SheetsID == "" {
			h.reply(replyToken, "尚未設定 GOOGLE_SHEETS_ID")
			return
		}
		url := "https://docs.google.com/spreadsheets/d/" + h.cfg.SheetsID + "/edit"
		h.reply(replyToken, "📊 成績試算表：\n"+url)

	case matchesAny(text, "/管理員列表", "/admins"):
		if !h.isAdmin(ctx, userID) {
			h.reply(replyToken, "⚠️ 只有管理員可以查看")
			return
		}
		admins, err := h.cfg.Store.ListAdmins(ctx)
		if err != nil {
			h.reply(replyToken, "❌ 查詢失敗："+err.Error())
			return
		}
		if len(admins) == 0 {
			h.reply(replyToken, "（Store 內沒有管理員；只靠環境變數 ADMIN_USER_IDS）")
			return
		}
		lines := []string{"👮 管理員列表："}
		for _, a := range admins {
			line := "• " + a.UserID
			if a.Name != "" {
				line = "• " + a.Name + "（" + a.UserID + "）"
			}
			lines = append(lines, line)
		}
		h.reply(replyToken, strings.Join(lines, "\n"))

	case matchesAny(text, "/賽事列表", "/list", "/tournaments"):
		list, err := h.cfg.Store.ListTournaments(ctx)
		if err != nil {
			h.reply(replyToken, "❌ 查詢失敗："+err.Error())
			return
		}
		if len(list) == 0 {
			h.reply(replyToken, "（目前沒有賽事）\n用「/新增賽事 名稱」新增")
			return
		}
		h.reply(replyToken, "📋 目前賽事列表：\n"+strings.Join(list, "\n"))

	case matchesAny(text, "/help", "/說明"):
		h.reply(replyToken, helpText)

	case text == "/dhelp":
		if !h.isAdmin(ctx, userID) {
			// 假裝不認識這個指令
			return
		}
		h.reply(replyToken, adminHelpText)

	default:
		// 可能是 /<賽事名稱> 查詢
		if strings.HasPrefix(text, "/") && !strings.Contains(text, " ") {
			name := strings.TrimPrefix(text, "/")
			if name != "" {
				h.maybeQueryTournament(ctx, replyToken, name)
			}
		}
		// 其他訊息不回
	}
}

const helpText = `🏸 羽球成績 Bot 指令

/登記、/add、/register
→ 取得登記連結（自己選賽事）

/所有連結、/all、/links
→ 列出所有賽事的專屬連結

/賽事列表、/list、/tournaments
→ 顯示目前所有賽事

/賽事名稱
→ 查詢該場比賽的得獎紀錄（例：/清晨盃）

/我的ID、/myid、/id
→ 顯示你的 LINE User ID

/help、/說明
→ 顯示這份說明`

const adminHelpText = `🔐 管理員指令

/開單 <賽事>、/open <賽事>
→ 發起成績登記（會自動加入列表）

/新增賽事 <賽事>、/addt <賽事>
→ 新增賽事到下拉

/刪除賽事 <賽事>、/delt <賽事>
→ 從下拉移除

/表單、/sheet
→ 取得 Google 試算表連結

/新增管理員 <Uxxx>、/addadmin <Uxxx>
→ 新增管理員

/刪除管理員 <Uxxx>、/deladmin <Uxxx>
→ 移除管理員

/管理員列表、/admins
→ 顯示所有管理員

/dhelp → 顯示這份說明`

// handleBootstrap 只有在 Store 沒有任何 admin 時才會真的執行，否則靜默。
func (h *Handler) handleBootstrap(ctx context.Context, replyToken, userID string) {
	if userID == "" {
		// 連 userID 都拿不到（可能是群組且使用者沒加 bot 為好友）
		h.reply(replyToken, "⚠️ 請先加 bot 為好友，再試一次")
		return
	}
	admins, err := h.cfg.Store.ListAdmins(ctx)
	if err != nil {
		log.Printf("list admins: %v", err)
		return // 靜默
	}
	if len(admins) > 0 {
		// 已有管理員，拒絕自助 bootstrap
		return
	}

	// 嘗試拿 LINE 顯示名稱（失敗無妨）
	name := ""
	if p, err := h.bot.GetProfile(userID); err == nil && p != nil {
		name = p.DisplayName
	}

	if err := h.cfg.Store.AddAdmin(ctx, store.Admin{
		UserID:  userID,
		Name:    name,
		AddedAt: time.Now(),
	}); err != nil {
		h.reply(replyToken, "❌ 新增失敗："+err.Error())
		return
	}
	msg := "✅ 你已成為管理員"
	if name != "" {
		msg += "：" + name
	}
	h.reply(replyToken, msg+"\n\n輸入 /dhelp 查看管理員指令")
}

// maybeQueryTournament 如果 name 是已存在的賽事，回傳該場比賽所有紀錄；否則靜默。
func (h *Handler) maybeQueryTournament(ctx context.Context, replyToken, name string) {
	tournaments, err := h.cfg.Store.ListTournaments(ctx)
	if err != nil {
		return
	}
	found := false
	for _, t := range tournaments {
		if t == name {
			found = true
			break
		}
	}
	if !found {
		return // 不是已註冊的賽事，靜默
	}

	results, err := h.cfg.Store.ListByTournament(ctx, name)
	if err != nil {
		h.reply(replyToken, "❌ 查詢失敗："+err.Error())
		return
	}
	if len(results) == 0 {
		h.reply(replyToken, "📋 "+name+"\n（目前沒有紀錄）")
		return
	}

	const maxLines = 40 // LINE 訊息長度保護
	lines := []string{"🏆 " + name + " 成績紀錄（" + strconv.Itoa(len(results)) + " 筆）", "───────────"}
	for i, r := range results {
		if i >= maxLines {
			lines = append(lines, fmt.Sprintf("...（另有 %d 筆，請至試算表檢視）", len(results)-maxLines))
			break
		}
		line := fmt.Sprintf("• %s｜%s %s", r.PlayerName, r.Event, r.Rank)
		if r.Category != "" {
			line += "（" + r.Category + "）"
		}
		if r.Partner != "" {
			line += " / 搭檔:" + r.Partner
		}
		lines = append(lines, line)
	}
	h.reply(replyToken, strings.Join(lines, "\n"))
}

func (h *Handler) replyWithLIFF(replyToken, tournament string) {
	if h.cfg.LIFFID == "" {
		h.reply(replyToken, "尚未設定 LIFF_ID，請先到 .env 補上")
		return
	}
	url := fmt.Sprintf("https://liff.line.me/%s?t=%s", h.cfg.LIFFID, tournament)
	msg := fmt.Sprintf("📋 %s 成績登記\n👉 %s\n\n（每人可填多項，可隨時回來修改）", tournament, url)
	h.reply(replyToken, msg)
}

func (h *Handler) reply(replyToken, text string) {
	_, err := h.bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
		ReplyToken: replyToken,
		Messages: []messaging_api.MessageInterface{
			messaging_api.TextMessage{Text: text},
		},
	})
	if err != nil {
		log.Printf("reply: %v", err)
	}
}

// ===== LIFF 表單後端 API =====

// GetOptions 提供下拉選單來源 + LIFF ID 給前端。
// 賽事從 Store 動態讀取（管理員透過 LINE 指令維護）。
func (h *Handler) GetOptions(c *gin.Context) {
	opts := domain.DefaultOptions()
	if stored, err := h.cfg.Store.ListTournaments(c.Request.Context()); err == nil {
		opts.Tournaments = stored
	} else {
		log.Printf("list tournaments failed: %v", err)
	}
	c.JSON(http.StatusOK, gin.H{
		"liff_id": h.cfg.LIFFID,
		"options": opts,
	})
}

type submitRequest struct {
	IDToken    string `json:"id_token"`  // LIFF 拿到的 ID token（未來驗證用）
	UserID     string `json:"user_id"`   // LINE userId
	UserName   string `json:"user_name"` // LINE 顯示名稱（輸入人）
	PlayerName string `json:"player_name"`
	Tournament string `json:"tournament"`
	Category   string `json:"category"`
	Event      string `json:"event"`
	Partner    string `json:"partner"`
	Rank       string `json:"rank"`
	Note       string `json:"note"`
}

func (h *Handler) SubmitResult(c *gin.Context) {
	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.UserID == "" || req.PlayerName == "" || req.Tournament == "" || req.Event == "" || req.Rank == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id / player_name / tournament / event / rank 必填"})
		return
	}

	r := domain.MatchResult{
		ID:          newID(),
		UserID:      req.UserID,
		UserName:    req.UserName,
		PlayerName:  req.PlayerName,
		Tournament:  req.Tournament,
		Category:    req.Category,
		Event:       req.Event,
		Partner:     req.Partner,
		Rank:        req.Rank,
		Note:        req.Note,
		SubmittedAt: time.Now(),
	}

	if err := h.cfg.Store.SaveResult(c.Request.Context(), r); err != nil {
		log.Printf("save result: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "id": r.ID})
}

func newID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// 預留：將來要把 ID token 驗證接上時用
var _ = json.Marshal
