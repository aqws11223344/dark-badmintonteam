// Package line 處理 LINE webhook 事件，並提供 LIFF 表單後端 API。
package line

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
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

	switch {
	case strings.HasPrefix(text, "/開單"):
		name := strings.TrimSpace(strings.TrimPrefix(text, "/開單"))
		if name == "" {
			h.reply(replyToken, "用法：/開單 2026清晨杯")
			return
		}
		h.replyWithLIFF(replyToken, name)

	case text == "/我的成績":
		h.reply(replyToken, "（此功能需在個人聊天使用，將於下一版加入）")

	case text == "/help" || text == "/說明":
		h.reply(replyToken, helpText)

	default:
		// 群組裡其他訊息不回，避免洗版
	}
}

const helpText = `🏸 羽球成績 Bot 用法：
/開單 賽事名稱   → 發起一場賽事的成績登記
/我的成績        → 查看你的成績紀錄
/說明            → 顯示這份說明`

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

// GetOptions 提供下拉選單來源給 LIFF。
func (h *Handler) GetOptions(c *gin.Context) {
	c.JSON(http.StatusOK, domain.DefaultOptions())
}

type submitRequest struct {
	IDToken    string `json:"id_token"` // LIFF 拿到的 ID token，用來驗使用者
	UserID     string `json:"user_id"`  // 退路（沒 verify 時使用）
	UserName   string `json:"user_name"`
	Tournament string `json:"tournament"`
	AgeGroup   string `json:"age_group"`
	Class      string `json:"class"`
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
	if req.UserID == "" || req.Tournament == "" || req.Event == "" || req.Rank == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id / tournament / event / rank 必填"})
		return
	}

	r := domain.MatchResult{
		ID:          newID(),
		UserID:      req.UserID,
		UserName:    req.UserName,
		Tournament:  req.Tournament,
		AgeGroup:    req.AgeGroup,
		Class:       req.Class,
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
