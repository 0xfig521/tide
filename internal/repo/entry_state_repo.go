package repo

import (
	"github.com/0xfig521/tide/internal/db"
	"github.com/0xfig521/tide/internal/models"
)

type EntryStateRepo struct {
	db *db.DB
}

func NewEntryStateRepo(db *db.DB) *EntryStateRepo {
	return &EntryStateRepo{db: db}
}

func (r *EntryStateRepo) SetState(entryID int64, state string) error {
	_, err := r.db.Conn.Exec(`
		INSERT OR REPLACE INTO entry_states (entry_id, state, updated_at)
		VALUES (?, ?, datetime('now'))
	`, entryID, state)
	return err
}

func (r *EntryStateRepo) SetStateWithTags(entryID int64, state string, tags string) error {
	_, err := r.db.Conn.Exec(`
		INSERT OR REPLACE INTO entry_states (entry_id, state, tags, updated_at)
		VALUES (?, ?, ?, datetime('now'))
	`, entryID, state, tags)
	return err
}

func (r *EntryStateRepo) SetStateFull(entryID int64, state string, tags string, note string) error {
	_, err := r.db.Conn.Exec(`
		INSERT OR REPLACE INTO entry_states (entry_id, state, tags, note, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'))
	`, entryID, state, tags, note)
	return err
}

func (r *EntryStateRepo) GetState(entryID int64) (*models.EntryState, error) {
	row := r.db.Conn.QueryRow(`
		SELECT entry_id, state, COALESCE(tags,''), COALESCE(note,''),
		       COALESCE(processed_at,''), created_at, updated_at
		FROM entry_states
		WHERE entry_id = ?
	`, entryID)
	es := &models.EntryState{}
	var processedAt, createdAt, updatedAt string
	if err := row.Scan(
		&es.EntryID, &es.State, &es.Tags, &es.Note,
		&processedAt, &createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}
	es.ProcessedAt = parseTime(processedAt)
	es.CreatedAt = mustParseTime(createdAt)
	es.UpdatedAt = mustParseTime(updatedAt)
	return es, nil
}

func (r *EntryStateRepo) ListByState(state string, limit, offset int) ([]*models.EntryState, error) {
	rows, err := r.db.Conn.Query(`
		SELECT entry_id, state, COALESCE(tags,''), COALESCE(note,''),
		       COALESCE(processed_at,''), created_at, updated_at
		FROM entry_states
		WHERE state = ?
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`, state, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []*models.EntryState
	for rows.Next() {
		es := &models.EntryState{}
		var processedAt, createdAt, updatedAt string
		if err := rows.Scan(
			&es.EntryID, &es.State, &es.Tags, &es.Note,
			&processedAt, &createdAt, &updatedAt,
		); err != nil {
			return nil, err
		}
		es.ProcessedAt = parseTime(processedAt)
		es.CreatedAt = mustParseTime(createdAt)
		es.UpdatedAt = mustParseTime(updatedAt)
		states = append(states, es)
	}
	return states, rows.Err()
}
