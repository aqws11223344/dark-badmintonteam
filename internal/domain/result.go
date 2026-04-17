package domain

import "time"

// MatchResult 一筆隊員的單一項目成績。
// 例：王小明 在 2026清晨杯 30歲組甲組 男雙（搭檔李大華）拿到金牌。
type MatchResult struct {
	ID          string    `json:"id,omitempty"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	Tournament  string    `json:"tournament"` // 例：2026清晨杯
	AgeGroup    string    `json:"age_group"`  // 例：30歲組
	Class       string    `json:"class"`      // 例：甲組
	Event       string    `json:"event"`      // 男單/女單/男雙/女雙/混雙
	Partner     string    `json:"partner,omitempty"`
	Rank        string    `json:"rank"` // 金/銀/銅/四強/八強/十六強/未晉級
	Note        string    `json:"note,omitempty"`
	SubmittedAt time.Time `json:"submitted_at"`
}

// Options LIFF 表單下拉選單來源。先寫死，未來可改成從 Sheet/DB 讀取。
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
		AgeGroups:   []string{"公開組", "30歲組", "40歲組", "50歲組", "60歲組"},
		Classes:     []string{"甲組", "乙組", "丙組"},
		Events:      []string{"男單", "女單", "男雙", "女雙", "混雙"},
		Ranks:       []string{"金牌", "銀牌", "銅牌", "四強", "八強", "十六強", "未晉級"},
	}
}
