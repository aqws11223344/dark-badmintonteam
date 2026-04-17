// Package dual 把寫入同時送到 primary 和 mirror，讀取一律走 primary。
// 用途：Turso 當主庫（bot 查詢快），Sheets 當鏡像（教練可看）。
package dual

import (
	"context"
	"log"

	"github.com/aqws11223344/dark-badmintonteam/internal/domain"
	"github.com/aqws11223344/dark-badmintonteam/internal/store"
)

type Store struct {
	primary store.Store
	mirror  store.Store
}

func New(primary, mirror store.Store) *Store {
	return &Store{primary: primary, mirror: mirror}
}

func (s *Store) SaveResult(ctx context.Context, r domain.MatchResult) error {
	if err := s.primary.SaveResult(ctx, r); err != nil {
		return err
	}
	// mirror 失敗只記 log，不阻擋主流程
	if err := s.mirror.SaveResult(ctx, r); err != nil {
		log.Printf("dual store: mirror save failed (id=%s): %v", r.ID, err)
	}
	return nil
}

func (s *Store) ListByUser(ctx context.Context, userID string) ([]domain.MatchResult, error) {
	return s.primary.ListByUser(ctx, userID)
}

func (s *Store) ListByTournament(ctx context.Context, t string) ([]domain.MatchResult, error) {
	return s.primary.ListByTournament(ctx, t)
}

func (s *Store) ListTournaments(ctx context.Context) ([]string, error) {
	return s.primary.ListTournaments(ctx)
}

func (s *Store) AddTournament(ctx context.Context, name string) error {
	if err := s.primary.AddTournament(ctx, name); err != nil {
		return err
	}
	if err := s.mirror.AddTournament(ctx, name); err != nil {
		log.Printf("dual store: mirror add tournament failed: %v", err)
	}
	return nil
}

func (s *Store) RemoveTournament(ctx context.Context, name string) error {
	if err := s.primary.RemoveTournament(ctx, name); err != nil {
		return err
	}
	if err := s.mirror.RemoveTournament(ctx, name); err != nil {
		log.Printf("dual store: mirror remove tournament failed: %v", err)
	}
	return nil
}

func (s *Store) ListAdmins(ctx context.Context) ([]store.Admin, error) {
	return s.primary.ListAdmins(ctx)
}

func (s *Store) AddAdmin(ctx context.Context, a store.Admin) error {
	if err := s.primary.AddAdmin(ctx, a); err != nil {
		return err
	}
	if err := s.mirror.AddAdmin(ctx, a); err != nil {
		log.Printf("dual store: mirror add admin failed: %v", err)
	}
	return nil
}

func (s *Store) RemoveAdmin(ctx context.Context, userID string) error {
	if err := s.primary.RemoveAdmin(ctx, userID); err != nil {
		return err
	}
	if err := s.mirror.RemoveAdmin(ctx, userID); err != nil {
		log.Printf("dual store: mirror remove admin failed: %v", err)
	}
	return nil
}

func (s *Store) Close() error {
	if err := s.primary.Close(); err != nil {
		return err
	}
	return s.mirror.Close()
}
