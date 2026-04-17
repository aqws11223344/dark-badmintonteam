package store

import (
	"context"
	"sort"
	"sync"

	"github.com/aqws11223344/dark-badmintonteam/internal/domain"
)

// Store 統一的儲存介面。Sheets / Turso / Memory 都實作這個。
type Store interface {
	SaveResult(ctx context.Context, r domain.MatchResult) error
	ListByUser(ctx context.Context, userID string) ([]domain.MatchResult, error)
	ListByTournament(ctx context.Context, tournament string) ([]domain.MatchResult, error)
	Close() error
}

// NewMemory 沒設定外部儲存時的 fallback，重啟即遺失。
func NewMemory() Store {
	return &memoryStore{data: make(map[string]domain.MatchResult)}
}

type memoryStore struct {
	mu   sync.RWMutex
	data map[string]domain.MatchResult
}

func (m *memoryStore) SaveResult(_ context.Context, r domain.MatchResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.ID == "" {
		r.ID = r.UserID + "|" + r.Tournament + "|" + r.AgeGroup + "|" + r.Class + "|" + r.Event
	}
	m.data[r.ID] = r
	return nil
}

func (m *memoryStore) ListByUser(_ context.Context, userID string) ([]domain.MatchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []domain.MatchResult
	for _, r := range m.data {
		if r.UserID == userID {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SubmittedAt.After(out[j].SubmittedAt) })
	return out, nil
}

func (m *memoryStore) ListByTournament(_ context.Context, t string) ([]domain.MatchResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []domain.MatchResult
	for _, r := range m.data {
		if r.Tournament == t {
			out = append(out, r)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UserName < out[j].UserName })
	return out, nil
}

func (m *memoryStore) Close() error { return nil }
