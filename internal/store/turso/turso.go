// Package turso 的 libsql driver 尚未編入此 build。
package turso

import (
	"context"
	"errors"

	"github.com/aqws11223344/dark-badmintonteam/internal/domain"
	"github.com/aqws11223344/dark-badmintonteam/internal/store"
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

func (s *Store) ListTournaments(_ context.Context) ([]string, error) {
	return nil, errors.New("turso disabled")
}

func (s *Store) AddTournament(_ context.Context, _ string) error {
	return errors.New("turso disabled")
}

func (s *Store) RemoveTournament(_ context.Context, _ string) error {
	return errors.New("turso disabled")
}

func (s *Store) ListAdmins(_ context.Context) ([]store.Admin, error) {
	return nil, errors.New("turso disabled")
}

func (s *Store) AddAdmin(_ context.Context, _ store.Admin) error {
	return errors.New("turso disabled")
}

func (s *Store) RemoveAdmin(_ context.Context, _ string) error {
	return errors.New("turso disabled")
}

func (s *Store) Close() error { return nil }
