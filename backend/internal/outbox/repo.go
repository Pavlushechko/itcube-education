package outbox

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{ db *pgxpool.Pool }

func New(db *pgxpool.Pool) *Repo { return &Repo{db: db} }

func (r *Repo) Add(ctx context.Context, aggregateType string, aggregateID uuid.UUID, eventType string, payload map[string]any) error {
	b, _ := json.Marshal(payload)
	_, err := r.db.Exec(ctx, `
		insert into outbox_events(id, aggregate_type, aggregate_id, event_type, payload)
		values ($1,$2,$3,$4,$5)
	`, uuid.New(), aggregateType, aggregateID, eventType, b)
	return err
}
