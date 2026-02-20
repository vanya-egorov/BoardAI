package agents

import (
	"BoardAI/internal/config"
	"BoardAI/internal/llm"
	"context"
)

type Role string

const (
	RoleStrategist Role = "strategist_optimist"
	RoleFinancier  Role = "financier"
	RoleAuditor    Role = "auditor_skeptic"
	RoleAnalyst    Role = "market_analyst"
	RoleModerator  Role = "moderator"
)

type Agent interface {
	Role() Role
	Model() string
	SystemPrompt() string
	Run(ctx context.Context, idea string) (string, error)
}

type baseAgent struct {
	role         Role
	model        string
	systemPrompt string
	client       *llm.Client
}

func NewAgentsFromConfig(cli *llm.Client, cfg *config.Config) map[Role]Agent {
	return map[Role]Agent{
		RoleStrategist: &baseAgent{
			role:         RoleStrategist,
			model:        cfg.ModelStrategist,
			systemPrompt: llm.SystemPromptStrategist,
			client:       cli,
		},
		RoleFinancier: &baseAgent{
			role:         RoleFinancier,
			model:        cfg.ModelFinancier,
			systemPrompt: llm.SystemPromptFinancier,
			client:       cli,
		},
		RoleAuditor: &baseAgent{
			role:         RoleAuditor,
			model:        cfg.ModelAuditor,
			systemPrompt: llm.SystemPromptAuditor,
			client:       cli,
		},
		RoleAnalyst: &baseAgent{
			role:         RoleAnalyst,
			model:        cfg.ModelAnalyst,
			systemPrompt: llm.SystemPromptAnalyst,
			client:       cli,
		},
		RoleModerator: &baseAgent{
			role:         RoleModerator,
			model:        cfg.ModelModerator,
			systemPrompt: llm.SystemPromptModerator,
			client:       cli,
		},
	}
}

func (a *baseAgent) Role() Role {
	return a.role
}

func (a *baseAgent) Model() string {
	return a.model
}

func (a *baseAgent) SystemPrompt() string {
	return a.systemPrompt
}

func (a *baseAgent) Run(ctx context.Context, idea string) (string, error) {
	return a.client.Chat(ctx, a.model, a.systemPrompt, idea)
}
