package bot

import (
	"BoardAI/internal/models"
	"BoardAI/internal/orchestrator"
	"BoardAI/internal/repository"
	"context"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	stateIdle         = "IDLE"
	stateWaitingQuery = "STATE_WAITING_QUERY"
	stateProcessing   = "STATE_PROCESSING"
	stateLastAnalysis = "STATE_LAST_ANALYSIS"
)

type StateManager struct {
	mu    sync.RWMutex
	state map[int64]string
}

func NewStateManager() *StateManager {
	return &StateManager{
		state: make(map[int64]string),
	}
}

func (s *StateManager) Get(userID int64) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.state[userID]; ok {
		return v
	}
	return stateIdle
}

func (s *StateManager) Set(userID int64, st string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state[userID] = st
}

type Handler struct {
	bot          *tgbotapi.BotAPI
	repo         repository.AnalysisRepository
	orchestrator *orchestrator.Orchestrator
	state        *StateManager
	lastAnalysis sync.Map
}

func NewHandler(
	bot *tgbotapi.BotAPI,
	repo repository.AnalysisRepository,
	orc *orchestrator.Orchestrator,
	state *StateManager,
) *Handler {
	return &Handler{
		bot:          bot,
		repo:         repo,
		orchestrator: orc,
		state:        state,
	}
}

func (h *Handler) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 600
	updates := h.bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Println("context cancelled, stopping updates loop")
			return nil
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			h.handleUpdate(ctx, update)
		}
	}
}

func (h *Handler) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message != nil {
		h.handleMessage(ctx, update.Message)
		return
	}
	if update.CallbackQuery != nil {
		h.handleCallbackQuery(ctx, update.CallbackQuery)
		return
	}
}

func (h *Handler) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	log.Printf("incoming message: user_id=%d text=%q", userID, msg.Text)

	switch msg.Text {
	case "Новый анализ", "🔄 Новый анализ":
		h.askForIdea(msg)
		return
	case "Мои анализы", "📋 Мои анализы":
		h.showHistory(ctx, msg)
		return
	}

	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			h.handleStart(msg)
		case "new":
			h.askForIdea(msg)
		case "list":
			h.showHistory(ctx, msg)
		case "cancel":
			h.state.Set(userID, stateIdle)
			h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Действие отменено."))
		default:
			h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Неизвестная команда. /new - новый анализ."))
		}
		return
	}

	state := h.state.Get(userID)
	if state == stateWaitingQuery {
		h.processIdea(ctx, msg)
	} else {
		resp := tgbotapi.NewMessage(msg.Chat.ID, "Нажмите кнопку «Новый анализ», чтобы начать.")
		resp.ReplyMarkup = buildMainKeyboard()
		h.bot.Send(resp)
	}
}

func (h *Handler) askForIdea(msg *tgbotapi.Message) {
	h.state.Set(msg.From.ID, stateWaitingQuery)
	resp := tgbotapi.NewMessage(msg.Chat.ID, "Опишите бизнес-идею подробно. Я запущу экспертный совет (займет 3-5 мин).")
	resp.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	h.bot.Send(resp)
}

func (h *Handler) handleStart(msg *tgbotapi.Message) {
	resp := tgbotapi.NewMessage(msg.Chat.ID, "👋Привет! Я Board AI Bot — мультиагентный совет директоров для бизнес-идей."+
		"Нажми кнопку «Новый анализ» или отправь команду /new, чтобы начать.")
	resp.ReplyMarkup = buildMainKeyboard()
	h.bot.Send(resp)
}

func (h *Handler) processIdea(ctx context.Context, msg *tgbotapi.Message) {
	userID := msg.From.ID
	idea := msg.Text

	if h.state.Get(userID) == stateProcessing {
		h.bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "Анализ уже идет, пожалуйста, подождите."))
		return
	}

	h.state.Set(userID, stateProcessing)
	waitMsg := tgbotapi.NewMessage(msg.Chat.ID, "⏳ Анализ запущен. Я пришлю результат, как только эксперты закончат...")
	sent, _ := h.bot.Send(waitMsg)

	go func() {
		analysisCtx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
		defer cancel()

		log.Printf("DEBUG: Starting RunAnalysis for user %d", userID)
		analysis, err := h.orchestrator.RunAnalysis(analysisCtx, idea, int64(userID))

		if err != nil {
			log.Printf("RunAnalysis error: %v", err)
			h.state.Set(userID, stateIdle)
			h.bot.Send(tgbotapi.NewEditMessageText(sent.Chat.ID, sent.MessageID, "⚠️ Ошибка анализа. Попробуйте позже."))
			return
		}

		h.lastAnalysis.Store(int64(userID), analysis)
		h.state.Set(userID, stateLastAnalysis)

		fullText := renderAnalysisMarkdown(analysis)

		log.Printf("DEBUG: Analysis complete, text length: %d", len(fullText))

		if len(fullText) < 4000 {
			edit := tgbotapi.NewEditMessageText(sent.Chat.ID, sent.MessageID, fullText)

			edit.ParseMode = "HTML"
			edit.ReplyMarkup = buildMainKeyboard()
			h.bot.Send(edit)
		} else {
			h.bot.Send(tgbotapi.NewDeleteMessage(sent.Chat.ID, sent.MessageID))

			for i := 0; i < len(fullText); i += 3900 {
				end := i + 3900
				if end > len(fullText) {
					end = len(fullText)
				}
				part := fullText[i:end]
				msg := tgbotapi.NewMessage(sent.Chat.ID, part)
				if i+3900 >= len(fullText) {
					msg.ReplyMarkup = buildMainKeyboard()
				}
				h.bot.Send(msg)
			}
		}
	}()
}

func (h *Handler) handleCallbackQuery(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	switch cq.Data {
	case callbackNewAnalysis:
		h.askForIdea(cq.Message)
	case callbackSaveAnalysis:
		h.saveLastAnalysis(ctx, cq)
	case callbackListHistory:
		h.showHistory(ctx, cq.Message)
	default:
	}

	_, _ = h.bot.Request(tgbotapi.NewCallback(cq.ID, ""))
}

func (h *Handler) saveLastAnalysis(ctx context.Context, cq *tgbotapi.CallbackQuery) {
	val, ok := h.lastAnalysis.Load(int64(cq.From.ID))
	if !ok {
		resp := tgbotapi.NewMessage(cq.Message.Chat.ID, "Нет последнего анализа для сохранения\\. Сначала запусти анализ\\.")
		h.bot.Send(resp)
		return
	}

	analysis, ok := val.(*models.Analysis)
	if !ok || analysis == nil {
		return
	}

	if err := h.repo.Create(ctx, analysis); err != nil {
		log.Printf("save analysis error: %v", err)
		resp := tgbotapi.NewMessage(cq.Message.Chat.ID, "Не удалось сохранить анализ в базу данных\\.")
		h.bot.Send(resp)
		return
	}

	resp := tgbotapi.NewMessage(cq.Message.Chat.ID, "Анализ успешно сохранен в базу данных ✅")
	h.bot.Send(resp)
}

func (h *Handler) showHistory(ctx context.Context, msg *tgbotapi.Message) {
	analyses, err := h.repo.List(ctx, 5, 0)
	if err != nil {
		log.Printf("list analyses error: %v", err)
		resp := tgbotapi.NewMessage(msg.Chat.ID, "Не удалось получить историю анализов\\.")
		h.bot.Send(resp)
		return
	}

	if len(analyses) == 0 {
		resp := tgbotapi.NewMessage(msg.Chat.ID, "История пуста\\. Сначала проведи анализ новой идеи\\.")
		h.bot.Send(resp)
		return
	}

	var text string
	text = "*Последние анализы:*\n\n"
	for _, a := range analyses {
		idea := a.IdeaText
		if len(idea) > 80 {
			idea = idea[:80] + "..."
		}
		text += "• " + escapeMarkdownV2(idea) + "\n"
	}

	resp := tgbotapi.NewMessage(msg.Chat.ID, text)
	h.bot.Send(resp)
}
