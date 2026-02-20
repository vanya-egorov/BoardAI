package models

import "time"

type Analysis struct {
	ID         int64  `db:"id" json:"id"`
	UserID     int64  `db:"user_id" json:"user_id"`
	IdeaText   string `db:"idea_text" json:"idea_text"`
	Strategist string `db:"strategist" json:"strategist"`
	Financier  string `db:"financier" json:"financier"`
	Auditor    string `db:"auditor" json:"auditor"`
	Analyst    string `db:"analyst" json:"analyst"`
	Moderator  string `db:"moderator" json:"moderator"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
