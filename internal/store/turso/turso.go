// Package turso 的 libsql driver 尚未編入此 build。
// 之後要啟用時：把 libsql-client-go 加回 go.mod，並恢復完整實作（見 git history）。
package turso

import (
	"context"
	"errors"

	"github.com/aqws11223344/dark-badmintonteam/internal/domain"
)

type Store struct{}

func New(_, _ string) (*Store, error) {
	return nil, errors.New("turso driver disabled in this build (sheets-only MVP)")
}

func (s *Store) SaveResult(_ context.Context, _ domain.MatchResult) error {
	return errors.New("turso disabled")
}

func (s *Store) ListByUser(_ context.Context, _ string) ([]domain.MatchResult, error) {
	return nil, errors.New("turso disabled")
}

func (s *Store) ListByTournament(_ context.Context, _ string) ([]domain.MatchResult, error) {
	return nil, errors.New("turso disabled")
}

func (s *Store) Close() error { return nil }
