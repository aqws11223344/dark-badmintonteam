package domain

import "time"

// MatchResult 一筆隊員的單一項目成績。
type MatchResult struct {
	ID          string    `json:"id,omitempty"`
	UserID      string    `json:"user_id"`     // LINE userId（給系統識別用）
	UserName    string    `json:"user_name"`   // LINE 顯示名稱（輸入人）
	PlayerName  string    `json:"player_name"` // 選手姓名（由隊員自行填寫，可能是綽號）
	Tournament  string    `json:"tournament"`  // 例：2026清晨杯
	AgeGroup    string    `json:"age_group"`   // 自由輸入，例：30歲組 / 29下 / 未分組
	Class       string    `json:"class"`       // 自由輸入，例：公開 / 社會 / 甲組
	Event       string    `json:"event"`       // 男單/女單/男雙/女雙/混雙
	Partner     string    `json:"partner,omitempty"`
	Rank        string    `json:"rank"` // 冠軍/亞軍/季軍/四強/八強/十六強/未晉級
	Note        string    `json:"note,omitempty"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// Options LIFF 表單的建議值（datalist 用）。
// 年齡組/級別是「建議值」不是「限定值」，隊員可以自行輸入。
type Options struct {
	Tournaments []string `json:"tournaments"`
	AgeGroups   []string `json:"age_groups"`
	Classes     []string `json:"classes"`
	Events      []string `json:"events"`
	Ranks       []string `json:"ranks"`
}

func DefaultOptions() Options {
	return Options{
		Tournaments: []string{"2026清晨杯", "2026春季聯賽", "2026縣長盃"},
		AgeGroups:   []string{"公開組", "30歲組", "40歲組", "50歲組", "60歲組", "29下組", "30上組", "未分組"},
		Classes:     []string{"公開", "社會", "青年", "大專", "甲組", "乙組", "丙組"},
		Events:      []string{"男單", "女單", "男雙", "女雙", "混雙"},
		Ranks:       []string{"冠軍", "亞軍", "季軍", "四強", "八強", "十六強", "未晉級"},
	}
}
