package store

import (
	"context"
	"sort"
	"sync"

	"github.com/aqws11223344/dark-badmintonteam/internal/domain"
)

// Store 統一的儲存介面。
type Store interface {
	SaveResult(ctx context.Context, r domain.MatchResult) error
	ListByUser(ctx context.Context, userID string) ([]domain.MatchResult, error)
	ListByTournament(ctx context.Context, tournament string) ([]domain.MatchResult, error)

	// 賽事管理
	ListTournaments(ctx context.Context) ([]string, error)
	AddTournament(ctx context.Context, name string) error
	RemoveTournament(ctx context.Context, name string) error

	Close() error
}

func NewMemory() Store {
	return &memoryStore{
		data:        make(map[string]domain.MatchResult),
		tournaments: make(map[string]bool),
	}
}

type memoryStore struct {
	mu          sync.RWMutex
	data        map[string]domain.MatchResult
	tournaments map[string]bool
}

func (m *memoryStore) SaveResult(_ context.Context, r domain.MatchResult) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r.ID == "" {
		r.ID = r.UserID + "|" + r.Tournament + "|" + r.Category + "|" + r.Event
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

func (m *memoryStore) ListTournaments(_ context.Context) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]string, 0, len(m.tournaments))
	for t := range m.tournaments {
		out = append(out, t)
	}
	sort.Strings(out)
	return out, nil
}

func (m *memoryStore) AddTournament(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tournaments[name] = true
	return nil
}

func (m *memoryStore) RemoveTournament(_ context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tournaments, name)
	return nil
}

func (m *memoryStore) Close() error { return nil }
