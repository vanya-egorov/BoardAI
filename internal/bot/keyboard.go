package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	callbackNewAnalysis  = "new_analysis"
	callbackSaveAnalysis = "save_analysis"
	callbackListHistory  = "list_history"
)

func buildMainKeyboard() *tgbotapi.InlineKeyboardMarkup {
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🆕 Новый анализ", callbackNewAnalysis),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💾 Сохранить анализ", callbackSaveAnalysis),
			tgbotapi.NewInlineKeyboardButtonData("📜 Мои анализы", callbackListHistory),
		),
	)
	return &kb
}
