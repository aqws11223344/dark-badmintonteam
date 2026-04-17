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
	ChannelSecret string
	ChannelToken  string
	LIFFID        string
	Store         store.Store
}

type Handler struct {
	cfg Config
	bot *messaging_api.MessagingApiAPI
}

func New(cfg Config) (*Handler, error) {
	bot, err := messaging_api.NewMessagingApiAPI(cfg.ChannelToken)
	if err != nil {
		return nil, err
	}
	return &Handler{cfg: cfg, bot: bot}, nil
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
				h.handleText(e.ReplyToken, msg.Text)
			}
		case webhook.FollowEvent:
			h.reply(e.ReplyToken, "歡迎加入！\n在群組裡輸入「/開單 賽事名稱」開始蒐集成績 🏸")
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleText(replyToken, text string) {
	text = strings.TrimSpace(text)
	ctx := context.Background()

	switch {
	case strings.HasPrefix(text, "/開單"):
		name := strings.TrimSpace(strings.TrimPrefix(text, "/開單"))
		if name == "" {
			h.reply(replyToken, "用法：/開單 2026清晨杯")
			return
		}
		// 自動加入賽事列表（若不存在）
		if err := h.cfg.Store.AddTournament(ctx, name); err != nil {
			log.Printf("add tournament on /開單 failed: %v", err)
		}
		h.replyWithLIFF(replyToken, name)

	case strings.HasPrefix(text, "/新增賽事"):
		name := strings.TrimSpace(strings.TrimPrefix(text, "/新增賽事"))
		if name == "" {
			h.reply(replyToken, "用法：/新增賽事 2026清晨杯")
			return
		}
		if err := h.cfg.Store.AddTournament(ctx, name); err != nil {
			h.reply(replyToken, "❌ 新增失敗："+err.Error())
			return
		}
		h.reply(replyToken, "✅ 已新增賽事：\n"+name)

	case strings.HasPrefix(text, "/刪除賽事"):
		name := strings.TrimSpace(strings.TrimPrefix(text, "/刪除賽事"))
		if name == "" {
			h.reply(replyToken, "用法：/刪除賽事 2026清晨杯")
			return
		}
		if err := h.cfg.Store.RemoveTournament(ctx, name); err != nil {
			h.reply(replyToken, "❌ 刪除失敗："+err.Error())
			return
		}
		h.reply(replyToken, "🗑 已刪除賽事：\n"+name)

	case text == "/賽事列表":
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

	case text == "/help" || text == "/說明":
		h.reply(replyToken, helpText)

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

const helpText = `🏸 羽球成績 Bot 指令：

/開單 賽事名稱    → 發起成績登記（會自動加入列表）
/新增賽事 XXX     → 新增賽事到表單下拉
/刪除賽事 XXX     → 從表單下拉移除
/賽事列表         → 顯示目前所有賽事
/賽事名稱         → 查詢該場比賽所有得獎紀錄
/說明             → 顯示這份說明

例：/清晨盃`

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
