package cli

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"hop/internal/i18n"
)

var (
	outputFrameStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#2C2C34")).Padding(0, 1)
	outputTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2A878")).Bold(true)
	outputSubStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#8A8A92"))
	outputSection    = lipgloss.NewStyle().Foreground(lipgloss.Color("#E0976A")).Bold(true)
	outputKeyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#DCDCDC")).Background(lipgloss.Color("#2C2C34")).Padding(0, 1)
	outputTextStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#DCDCDC"))
	outputHintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#7A7A82"))
	outputCodeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2A878")).Background(lipgloss.Color("#2C2C34")).Padding(0, 1)
	outputOkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7FB582")).Bold(true)
	outputInfoStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#DCDCDC"))
	outputErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75")).Bold(true)
)

func RenderHelp(translator i18n.Translator) string {
	lines := []string{
		outputTitleStyle.Render(translator.T(i18n.MsgHelpTitle)),
		outputSubStyle.Render(translator.T(i18n.MsgHelpSubtitle)),
		"",
		outputSection.Render(translator.T(i18n.MsgHelpUsageHeader)),
		outputTextStyle.Render(translator.T(i18n.MsgHelpUsage)),
		"",
		outputSection.Render(translator.T(i18n.MsgHelpFlags)),
		helpFlagRow("--language en|ru", translator.T(i18n.MsgHelpLanguageFlag)),
		helpFlagRow("--path add|remove", translator.T(i18n.MsgHelpPathFlag)),
		helpFlagRow("-h, --help", translator.T(i18n.MsgHelpHelpFlag)),
		"",
		outputSection.Render(translator.T(i18n.MsgPathSessionHint)),
		outputHintStyle.Render(translator.T(i18n.MsgPathSessionCommand)),
	}
	return outputFrameStyle.Render(strings.Join(lines, "\n"))
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
		outputTitleStyle.Render(translator.T(i18n.MsgHelpTitle)),
		outputOkStyle.Render(status),
		outputCodeStyle.Render(resultPath),
	}
	if changed {
		lines = append(lines,
			"",
			outputSection.Render(translator.T(i18n.MsgPathSessionHint)),
			outputHintStyle.Render(translator.T(i18n.MsgPathSessionCommand)),
		)
	}
	return outputFrameStyle.Render(strings.Join(lines, "\n"))
}

func RenderError(translator i18n.Translator, err error) string {
	lines := []string{
		outputTitleStyle.Render(translator.T(i18n.MsgHelpTitle)),
		outputErrorStyle.Render(translator.Error(err)),
	}
	return outputFrameStyle.Render(strings.Join(lines, "\n"))
}

func helpFlagRow(flag, description string) string {
	return outputKeyStyle.Render(flag) + "  " + outputTextStyle.Render(description)
}
