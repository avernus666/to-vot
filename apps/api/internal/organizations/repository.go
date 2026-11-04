package organizations

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Organization struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Role            string    `json:"role"`
	DefaultCurrency string    `json:"default_currency"`
	Timezone        string    `json:"timezone"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListForUser(ctx context.Context, userID uuid.UUID) ([]Organization, error) {
	rows, err := r.db.Query(ctx, `
		SELECT o.id, o.name, o.slug, m.role, o.default_currency, o.timezone
		FROM organizations o
		JOIN organization_members m ON m.organization_id = o.id
		WHERE m.user_id = $1
		ORDER BY o.created_at ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	defer rows.Close()

	items := make([]Organization, 0)
	for rows.Next() {
		var item Organization
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Role, &item.DefaultCurrency, &item.Timezone); err != nil {
			return nil, fmt.Errorf("scan organization: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate organizations: %w", err)
	}
	return items, nil
}
