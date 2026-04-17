// Package sheets 把成績寫入 Google Sheets，方便教練/隊長直接看試算表。
//
// 分頁設計：每年一個分頁（例：2026、2027），自動建立。
// 欄位：
//   A: ID | B: SubmittedAt | C: UserID | D: UserName(LINE 輸入人) | E: PlayerName(選手姓名)
//   F: Tournament | G: AgeGroup | H: Class | I: Event | J: Partner | K: Rank | L: Note
package sheets

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"

	"github.com/aqws11223344/dark-badmintonteam/internal/domain"
)

const colRange = "A:L" // 12 欄

type Store struct {
	svc     *sheets.Service
	sheetID string

	mu    sync.Mutex
	known map[string]bool
}

func New(ctx context.Context, sheetID, credentialsFile string) (*Store, error) {
	var opts []option.ClientOption
	if credentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialsFile))
	}
	svc, err := sheets.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("new sheets service: %w", err)
	}

	s := &Store{
		svc:     svc,
		sheetID: sheetID,
		known:   make(map[string]bool),
	}
	if err := s.ensureSheet(ctx, yearTab(time.Now())); err != nil {
		return nil, fmt.Errorf("ensure current-year sheet: %w", err)
	}
	return s, nil
}

func yearTab(t time.Time) string {
	return strconv.Itoa(t.Year())
}

func (s *Store) ensureSheet(ctx context.Context, name string) error {
	s.mu.Lock()
	if s.known[name] {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	meta, err := s.svc.Spreadsheets.Get(s.sheetID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("get spreadsheet: %w", err)
	}

	exists := false
	for _, sh := range meta.Sheets {
		if sh.Properties.Title == name {
			exists = true
			break
		}
	}

	if !exists {
		_, err := s.svc.Spreadsheets.BatchUpdate(s.sheetID, &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{{
				AddSheet: &sheets.AddSheetRequest{
					Properties: &sheets.SheetProperties{Title: name},
				},
			}},
		}).Context(ctx).Do()
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("add sheet %q: %w", name, err)
		}
	}

	if err := s.ensureHeader(ctx, name); err != nil {
		return err
	}

	s.mu.Lock()
	s.known[name] = true
	s.mu.Unlock()
	return nil
}

func (s *Store) ensureHeader(ctx context.Context, tab string) error {
	headerRange := fmt.Sprintf("%s!A1:L1", tab)
	resp, err := s.svc.Spreadsheets.Values.Get(s.sheetID, headerRange).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("get header: %w", err)
	}
	if len(resp.Values) > 0 && len(resp.Values[0]) > 0 {
		return nil
	}
	header := []any{"ID", "SubmittedAt", "UserID", "UserName", "PlayerName",
		"Tournament", "AgeGroup", "Class", "Event", "Partner", "Rank", "Note"}
	_, err = s.svc.Spreadsheets.Values.Update(s.sheetID, headerRange, &sheets.ValueRange{
		Values: [][]any{header},
	}).ValueInputOption("RAW").Context(ctx).Do()
	return err
}

func (s *Store) SaveResult(ctx context.Context, r domain.MatchResult) error {
	tab := yearTab(r.SubmittedAt)
	if err := s.ensureSheet(ctx, tab); err != nil {
		return err
	}
	row := []any{
		r.ID,
		r.SubmittedAt.Format("2006-01-02 15:04:05"),
		r.UserID, r.UserName, r.PlayerName,
		r.Tournament, r.AgeGroup, r.Class, r.Event,
		r.Partner, r.Rank, r.Note,
	}
	appendArea := fmt.Sprintf("%s!%s", tab, colRange)
	_, err := s.svc.Spreadsheets.Values.Append(s.sheetID, appendArea, &sheets.ValueRange{
		Values: [][]any{row},
	}).ValueInputOption("RAW").InsertDataOption("INSERT_ROWS").Context(ctx).Do()
	return err
}

func (s *Store) ListByUser(ctx context.Context, userID string) ([]domain.MatchResult, error) {
	all, err := s.readAll(ctx)
	if err != nil {
		return nil, err
	}
	var out []domain.MatchResult
	for _, r := range all {
		if r.UserID == userID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *Store) ListByTournament(ctx context.Context, t string) ([]domain.MatchResult, error) {
	all, err := s.readAll(ctx)
	if err != nil {
		return nil, err
	}
	var out []domain.MatchResult
	for _, r := range all {
		if r.Tournament == t {
			out = append(out, r)
		}
	}
	return out, nil
}

func (s *Store) readAll(ctx context.Context) ([]domain.MatchResult, error) {
	meta, err := s.svc.Spreadsheets.Get(s.sheetID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get spreadsheet: %w", err)
	}

	var out []domain.MatchResult
	for _, sh := range meta.Sheets {
		title := sh.Properties.Title
		if !isYearTab(title) {
			continue
		}
		readRange := fmt.Sprintf("%s!A2:L", title)
		resp, err := s.svc.Spreadsheets.Values.Get(s.sheetID, readRange).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", title, err)
		}
		for _, row := range resp.Values {
			out = append(out, rowToResult(row))
		}
	}
	return out, nil
}

func isYearTab(s string) bool {
	if len(s) != 4 {
		return false
	}
	n, err := strconv.Atoi(s)
	return err == nil && n >= 2000 && n <= 2999
}

func rowToResult(row []any) domain.MatchResult {
	get := func(i int) string {
		if i >= len(row) {
			return ""
		}
		s, _ := row[i].(string)
		return s
	}
	return domain.MatchResult{
		ID:         get(0),
		UserID:     get(2),
		UserName:   get(3),
		PlayerName: get(4),
		Tournament: get(5),
		AgeGroup:   get(6),
		Class:      get(7),
		Event:      get(8),
		Partner:    get(9),
		Rank:       get(10),
		Note:       get(11),
	}
}

func (s *Store) Close() error { return nil }
