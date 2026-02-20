package repository

import (
	"BoardAI/internal/models"
	"context"
	"database/sql"
	"fmt"
)

type AnalysisRepository interface {
	Create(ctx context.Context, a *models.Analysis) error
	Get(ctx context.Context, id int64) (*models.Analysis, error)
	List(ctx context.Context, limit, offset int) ([]*models.Analysis, error)
}

type analysisRepository struct {
	db *sql.DB
}

func NewAnalysisRepository(db *sql.DB) AnalysisRepository {
	return &analysisRepository{db: db}
}

func (r *analysisRepository) Create(ctx context.Context, a *models.Analysis) error {
	query := `
		INSERT INTO analyses (
			user_id,
			idea_text,
			strategist,
			financier,
			auditor,
			analyst,
			moderator
		) VALUES ($1, $2, $3::jsonb, $4::jsonb, $5::jsonb, $6::jsonb, $7::jsonb)
		RETURNING id, created_at
	`

	strategistJSON := fmt.Sprintf(`{"role":"strategist","content":%q}`, a.Strategist)
	financierJSON := fmt.Sprintf(`{"role":"financier","content":%q}`, a.Financier)
	auditorJSON := fmt.Sprintf(`{"role":"auditor","content":%q}`, a.Auditor)
	analystJSON := fmt.Sprintf(`{"role":"analyst","content":%q}`, a.Analyst)
	moderatorJSON := fmt.Sprintf(`{"role":"moderator","content":%q}`, a.Moderator)

	err := r.db.QueryRowContext(
		ctx,
		query,
		a.UserID,
		a.IdeaText,
		strategistJSON,
		financierJSON,
		auditorJSON,
		analystJSON,
		moderatorJSON,
	).Scan(&a.ID, &a.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert analysis: %w", err)
	}

	return nil
}

func (r *analysisRepository) Get(ctx context.Context, id int64) (*models.Analysis, error) {
	query := `
		SELECT
			id,
			user_id,
			idea_text,
			COALESCE(strategist::text, '') AS strategist,
			COALESCE(financier::text, '') AS financier,
			COALESCE(auditor::text, '')   AS auditor,
			COALESCE(analyst::text, '')   AS analyst,
			COALESCE(moderator::text, '') AS moderator,
			created_at
		FROM analyses
		WHERE id = $1
	`

	row := r.db.QueryRowContext(ctx, query, id)
	var a models.Analysis
	if err := row.Scan(
		&a.ID,
		&a.UserID,
		&a.IdeaText,
		&a.Strategist,
		&a.Financier,
		&a.Auditor,
		&a.Analyst,
		&a.Moderator,
		&a.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get analysis: %w", err)
	}

	return &a, nil
}

func (r *analysisRepository) List(ctx context.Context, limit, offset int) ([]*models.Analysis, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT
			id,
			user_id,
			idea_text,
			COALESCE(strategist::text, '') AS strategist,
			COALESCE(financier::text, '') AS financier,
			COALESCE(auditor::text, '')   AS auditor,
			COALESCE(analyst::text, '')   AS analyst,
			COALESCE(moderator::text, '') AS moderator,
			created_at
		FROM analyses
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list analyses: %w", err)
	}
	defer rows.Close()

	var result []*models.Analysis
	for rows.Next() {
		var a models.Analysis
		if err := rows.Scan(
			&a.ID,
			&a.UserID,
			&a.IdeaText,
			&a.Strategist,
			&a.Financier,
			&a.Auditor,
			&a.Analyst,
			&a.Moderator,
			&a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan analysis: %w", err)
		}
		result = append(result, &a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return result, nil
}
