package domain

import "time"

// MatchResult 一筆隊員的單一項目成績。
type MatchResult struct {
	ID          string    `json:"id,omitempty"`
	UserID      string    `json:"user_id"`     // LINE userId（系統識別用）
	UserName    string    `json:"user_name"`   // LINE 顯示名稱（輸入人）
	PlayerName  string    `json:"player_name"` // 選手姓名（由隊員自行填寫）
	Tournament  string    `json:"tournament"`  // 例：2026清晨杯
	Category    string    `json:"category"`    // 組別（合併年齡+級別）
	Event       string    `json:"event"`       // 男單/女單/男雙/女雙/混雙/團體
	Partner     string    `json:"partner,omitempty"`
	Rank        string    `json:"rank"` // 冠軍/亞軍/季軍/四強/八強/十六強/未晉級
	Note        string    `json:"note,omitempty"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// Options LIFF 表單的建議值（datalist 用）。使用者可自行輸入非建議值。
type Options struct {
	Tournaments []string `json:"tournaments"`
	Categories  []string `json:"categories"`
	Events      []string `json:"events"`
	Ranks       []string `json:"ranks"`
}

func DefaultOptions() Options {
	return Options{
		Tournaments: []string{"2026清晨杯", "2026春季聯賽", "2026縣長盃"},
		Categories:  []string{"公開組", "青年組", "大專組", "甲組", "乙組"},
		Events:      []string{"男單", "女單", "男雙", "女雙", "混雙", "團體"},
		Ranks:       []string{"冠軍", "亞軍", "季軍", "四強", "八強", "十六強", "未晉級"},
	}
}
