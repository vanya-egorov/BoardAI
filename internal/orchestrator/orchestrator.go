package orchestrator

import (
	"BoardAI/internal/agents"
	"BoardAI/internal/config"
	"BoardAI/internal/llm"
	"BoardAI/internal/models"
	"context"
	"fmt"
	"time"
)

// Orchestrator runs multiple agents and aggregates results.
type Orchestrator struct {
	agents map[agents.Role]agents.Agent
	cfg    *config.Config
}

// NewOrchestrator constructs orchestrator with all agents from LLM client.
func NewOrchestrator(cli *llm.Client, cfg *config.Config) *Orchestrator {
	return &Orchestrator{
		agents: agents.NewAgentsFromConfig(cli, cfg),
		cfg:    cfg,
	}
}

// limitText обрезает слишком длинные ответы экспертов, чтобы не переполнять контекст Ollama.
func limitText(s string, maxChars int) string {
	if len(s) <= maxChars {
		return s
	}
	return s[:maxChars] + "... [текст сокращен]"
}

// RunAnalysis runs expert agents sequentially to save CPU resources.
func (o *Orchestrator) RunAnalysis(parentCtx context.Context, idea string, userID int64) (*models.Analysis, error) {
	if o.agents == nil {
		return nil, fmt.Errorf("agents not initialized")
	}

	// Увеличиваем таймаут до 15 минут, так как 5 агентов на CPU — это долго
	ctx, cancel := context.WithTimeout(parentCtx, 15*time.Minute)
	defer cancel()

	var strategist, financier, auditor, analyst string
	var err error

	// 1. Стратег
	if ag, ok := o.agents[agents.RoleStrategist]; ok {
		strategist, err = ag.Run(ctx, idea)
		if err != nil {
			fmt.Printf("Strategist error: %v\n", err)
			strategist = "Ошибка анализа стратега"
		}
	}

	// 2. Финансист
	if ag, ok := o.agents[agents.RoleFinancier]; ok {
		financier, err = ag.Run(ctx, idea)
		if err != nil {
			fmt.Printf("Financier error: %v\n", err)
			financier = "Ошибка финансового анализа"
		}
	}

	// 3. Аудитор
	if ag, ok := o.agents[agents.RoleAuditor]; ok {
		auditor, err = ag.Run(ctx, idea)
		if err != nil {
			fmt.Printf("Auditor error: %v\n", err)
			auditor = "Ошибка аудита"
		}
	}

	// 4. Аналитик
	if ag, ok := o.agents[agents.RoleAnalyst]; ok {
		analyst, err = ag.Run(ctx, idea)
		if err != nil {
			fmt.Printf("Analyst error: %v\n", err)
			analyst = "Ошибка анализа рынка"
		}
	}

	// 5. Модератор (Финальный вердикт)
	moderatorAgent := o.agents[agents.RoleModerator]
	if moderatorAgent == nil {
		return nil, fmt.Errorf("moderator agent not initialized")
	}

	moderatorPrompt := fmt.Sprintf(
		"🏁 ФИНАЛЬНЫЙ ВЕРДИКТ\n\n"+
			"%s\n\n"+
			"--------------------------\n"+
			"💡 ИДЕЯ: %s\n\n"+
			"📋 КРАТКИЕ ОТЧЕТЫ ЭКСПЕРТОВ:\n"+
			"🔹 Стратег: %s\n"+
			"🔹 Финансист: %s\n"+
			"🔹 *Аудитор:* %s\n"+
			"🔹 *Аналитик:* %s",
		idea,
		limitText(strategist, 200),
		limitText(financier, 200),
		limitText(auditor, 200),
		limitText(analyst, 200),
	)

	moderator, err := moderatorAgent.Run(ctx, moderatorPrompt)
	if err != nil {
		return nil, fmt.Errorf("moderator run error: %w", err)
	}

	return &models.Analysis{
		UserID:     userID,
		IdeaText:   idea,
		Strategist: strategist,
		Financier:  financier,
		Auditor:    auditor,
		Analyst:    analyst,
		Moderator:  moderator,
	}, nil
}
