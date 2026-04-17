// Package turso 用 libsql 連 Turso（雲端 SQLite），給 bot 自己查詢用。
package turso

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	_ "github.com/tursodatabase/libsql-client-go/libsql"

	"github.com/aqws11223344/dark-badmintonteam/internal/domain"
)

const schema = `
CREATE TABLE IF NOT EXISTS results (
    id            TEXT PRIMARY KEY,
    submitted_at  TEXT NOT NULL,
    user_id       TEXT NOT NULL,
    user_name     TEXT NOT NULL,
    player_name   TEXT NOT NULL,
    tournament    TEXT NOT NULL,
    age_group     TEXT,
    class         TEXT,
    event         TEXT NOT NULL,
    partner       TEXT,
    rank          TEXT NOT NULL,
    note          TEXT
);
CREATE INDEX IF NOT EXISTS idx_results_user       ON results(user_id);
CREATE INDEX IF NOT EXISTS idx_results_tournament ON results(tournament);
CREATE INDEX IF NOT EXISTS idx_results_player     ON results(player_name);
`

type Store struct {
	db *sql.DB
}

func New(dbURL, authToken string) (*Store, error) {
	dsn := dbURL
	if authToken != "" {
		u, err := url.Parse(dbURL)
		if err != nil {
			return nil, fmt.Errorf("parse turso url: %w", err)
		}
		q := u.Query()
		q.Set("authToken", authToken)
		u.RawQuery = q.Encode()
		dsn = u.String()
	}

	db, err := sql.Open("libsql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open libsql: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) SaveResult(ctx context.Context, r domain.MatchResult) error {
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO results (id, submitted_at, user_id, user_name, player_name, tournament, age_group, class, event, partner, rank, note)
        VALUES (?,?,?,?,?,?,?,?,?,?,?,?)
        ON CONFLICT(id) DO UPDATE SET
            submitted_at = excluded.submitted_at,
            user_name    = excluded.user_name,
            player_name  = excluded.player_name,
            partner      = excluded.partner,
            rank         = excluded.rank,
            note         = excluded.note`,
		r.ID, r.SubmittedAt.Format("2006-01-02 15:04:05"),
		r.UserID, r.UserName, r.PlayerName,
		r.Tournament, r.AgeGroup, r.Class, r.Event,
		r.Partner, r.Rank, r.Note,
	)
	return err
}

const selectCols = `id, submitted_at, user_id, user_name, player_name, tournament, age_group, class, event, partner, rank, note`

func (s *Store) ListByUser(ctx context.Context, userID string) ([]domain.MatchResult, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+selectCols+` FROM results WHERE user_id = ? ORDER BY submitted_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scan(rows)
}

func (s *Store) ListByTournament(ctx context.Context, t string) ([]domain.MatchResult, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+selectCols+` FROM results WHERE tournament = ? ORDER BY player_name`, t)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scan(rows)
}

func scan(rows *sql.Rows) ([]domain.MatchResult, error) {
	var out []domain.MatchResult
	for rows.Next() {
		var r domain.MatchResult
		var submittedAt, ageGroup, class, partner, note sql.NullString
		if err := rows.Scan(&r.ID, &submittedAt, &r.UserID, &r.UserName, &r.PlayerName,
			&r.Tournament, &ageGroup, &class, &r.Event,
			&partner, &r.Rank, &note); err != nil {
			return nil, err
		}
		r.AgeGroup = ageGroup.String
		r.Class = class.String
		r.Partner = partner.String
		r.Note = note.String
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) Close() error { return s.db.Close() }
