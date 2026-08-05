package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Store struct {
	conn *pgx.Conn
}

func Connect(ctx context.Context, databaseURL string) (*Store, error) {
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	return &Store{conn: conn}, nil
}

func (s *Store) Close(ctx context.Context) {
	_ = s.conn.Close(ctx)
}

// ExistsByTitle replicates the rake task's deduplication strategy exactly.
func (s *Store) ExistsByTitle(ctx context.Context, title string) (bool, error) {
	var exists bool
	err := s.conn.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM sismos WHERE title = $1)`,
		title,
	).Scan(&exists)
	return exists, err
}

type SismoRecord struct {
	Title      string
	URL        string
	Place      string
	MagType    string
	Mag        float64
	Latitude   float64
	Longitude  float64
	Tsunami    bool
	ExternalID string
}

// Insert persists a new sismo row. Rails creates the magType identifier with
// mixed case, so PostgreSQL requires it to be quoted here.
func (s *Store) Insert(ctx context.Context, r SismoRecord) error {
	_, err := s.conn.Exec(ctx, `
		INSERT INTO sismos (title, url, place, "magType", mag, latitude, longitude, tsunami, external_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
	`, r.Title, r.URL, r.Place, r.MagType, r.Mag, r.Latitude, r.Longitude, r.Tsunami, r.ExternalID)
	return err
}
