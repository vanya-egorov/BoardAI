package bot

import (
	"BoardAI/internal/models"
	"fmt"
	"strings"
)

func escapeMarkdownV2(text string) string {
	replacer := strings.NewReplacer(
		`_`, `\_`,
		`*`, `\*`,
		`[`, `\[`,
		`]`, `\]`,
		`(`, `\(`,
		`)`, `\)`,
		`~`, `\~`,
		"`", "\\`",
		`>`, `\>`,
		`#`, `\#`,
		`+`, `\+`,
		`-`, `\-`,
		`=`, `\=`,
		`|`, `\|`,
		`{`, `\{`,
		`}`, `\}`,
		`.`, `\.`,
		`!`, `\!`,
	)
	return replacer.Replace(text)
}

func renderAnalysisMarkdown(a *models.Analysis) string {
	return fmt.Sprintf(
		"📊 РЕЗУЛЬТАТЫ АНАЛИЗА\n\n"+
			"💡 ИДЕЯ: %s\n\n"+
			"👨‍💼 ВЕРДИКТ МОДЕРАТОРА:\n%s\n\n"+
			"📈 СТРАТЕГИЯ:\n%s\n\n"+
			"💰 ФИНАНСЫ:\n%s\n",
		a.IdeaText, a.Moderator, a.Strategist, a.Financier,
	)
}
