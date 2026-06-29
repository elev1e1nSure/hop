package cli

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/elev1e1nSure/hop/internal/i18n"
)

var (
	titleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2A878")).Bold(true)
	sectionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E0976A")).Bold(true)
	keyStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#DCDCDC"))
	textStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#DCDCDC"))
	hintStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7A7A82"))
	okStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#7FB582")).Bold(true)
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75")).Bold(true)
	codeStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2A878"))
)

func RenderHelp(translator i18n.Translator) string {
	lines := []string{
		titleStyle.Render(translator.T(i18n.MsgHelpTitle)),
		hintStyle.Render(translator.T(i18n.MsgHelpSubtitle)),
		"",
		sectionStyle.Render(translator.T(i18n.MsgHelpUsageHeader)),
		textStyle.Render(translator.T(i18n.MsgHelpUsage)),
		"",
		sectionStyle.Render(translator.T(i18n.MsgHelpFlags)),
		helpFlagRow("--language en|ru", translator.T(i18n.MsgHelpLanguageFlag)),
		helpFlagRow("--path add|remove", translator.T(i18n.MsgHelpPathFlag)),
		helpFlagRow("-h, --help", translator.T(i18n.MsgHelpHelpFlag)),
	}
	return strings.Join(lines, "\n")
}

func RenderPathResult(translator i18n.Translator, action PathAction, resultPath string, changed bool) string {
	status := translator.T(i18n.MsgPathAddChanged)
	if action == PathActionRemove {
		status = translator.T(i18n.MsgPathRemoveChanged)
	}
	if !changed {
		status = translator.T(i18n.MsgPathAddUnchanged)
		if action == PathActionRemove {
			status = translator.T(i18n.MsgPathRemoveUnchanged)
		}
	}

	lines := []string{
		titleStyle.Render(translator.T(i18n.MsgHelpTitle)),
	}
	if changed {
		lines = append(lines,
			okStyle.Render(status),
			textStyle.Render(resultPath),
			"",
			sectionStyle.Render(translator.T(i18n.MsgPathSessionHint)),
			codeStyle.Render(translator.T(i18n.MsgPathSessionCommand)),
		)
		return strings.Join(lines, "\n")
	}

	lines = append(lines, textStyle.Render(status))
	if resultPath != "" {
		lines = append(lines, hintStyle.Render(resultPath))
	}
	return strings.Join(lines, "\n")
}

func RenderError(translator i18n.Translator, err error) string {
	return strings.Join([]string{
		titleStyle.Render(translator.T(i18n.MsgHelpTitle)),
		errStyle.Render(translator.Error(err)),
	}, "\n")
}

func helpFlagRow(flag, description string) string {
	return keyStyle.Render(flag) + "  " + textStyle.Render(description)
}
