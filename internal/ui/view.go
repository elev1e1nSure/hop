package ui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"hop/internal/domain"
	"hop/internal/i18n"
	"hop/internal/util"
)

func (model Model) View() string {
	switch model.mode {
	case modeForm:
		return model.formView()
	case modeConfirmDelete:
		return model.confirmView()
	case modeConnecting:
		return model.connectingView()
	default:
		return model.browseView()
	}
}

func (model Model) headerView() string {
	style1 := lipgloss.NewStyle().Foreground(lipgloss.Color("#F2A878")).Bold(true)
	style2 := lipgloss.NewStyle().Foreground(lipgloss.Color("#E0976A")).Bold(true)
	style3 := lipgloss.NewStyle().Foreground(lipgloss.Color("#C9A48E")).Bold(true)
	return style1.Render("  ╦ ╦ ╔═╗ ╔═╗") + "\n" +
		style2.Render("  ╠═╣ ║ ║ ╠═╝") + "\n" +
		style3.Render("  ╩ ╩ ╚═╝ ╩  ")
}

func (model Model) browseView() string {
	filterView := model.filter.View()
	padding := util.Max(0, model.width-4-lipgloss.Width(filterView))
	filterView += lipgloss.NewStyle().Background(filterBackground).Render(strings.Repeat(" ", padding))
	divider := "  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#2C2C34")).Render(strings.Repeat("─", util.Max(0, model.width-4)))

	parts := []string{
		"",
		model.headerView(),
		"",
		divider,
		"",
		"  " + filterBoxStyle.Render(filterView),
		"",
	}
	if server, ok := model.selectedServer(); ok {
		if model.showDetailsPanel() {
			parts = append(parts, model.listWithDetailsView(server), "")
		} else {
			parts = append(parts, model.list.View(), "", model.compactDetailsView(server), "")
		}
	} else {
		parts = append(parts, model.list.View(), "")
	}
	parts = append(parts, divider, "")
	if model.errorText != "" {
		parts = append(parts, "  "+accentStyle.Render(util.Truncate(model.errorText, util.Max(0, model.width-2))), "")
	}
	parts = append(parts, model.footerView())
	return strings.Join(parts, "\n")
}

func (model Model) listWithDetailsView(server domain.Server) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		model.list.View(),
		strings.Repeat(" ", detailsPanelGap),
		model.detailsPanelView(server),
	)
}

func (model Model) detailsPanelView(server domain.Server) string {
	const labelWidth = 10
	valueWidth := detailsPanelWidth - labelWidth - 6
	lines := []string{
		panelTitleStyle.Render(util.Truncate(util.Sanitize(server.Alias), detailsPanelWidth-4)),
		"",
		model.detailsRow(i18n.MsgDetailsHost, server.Host, labelWidth, valueWidth),
		model.detailsRow(i18n.MsgDetailsUser, optionalValue(model.translator, server.User), labelWidth, valueWidth),
		model.detailsRow(i18n.MsgDetailsPort, strconv.Itoa(server.Port), labelWidth, valueWidth),
		model.detailsRow(i18n.MsgDetailsIdentity, optionalValue(model.translator, server.IdentityFile), labelWidth, valueWidth),
		model.detailsRow(i18n.MsgDetailsProxy, proxyValue(model.translator, server), labelWidth, valueWidth),
		model.detailsRow(i18n.MsgDetailsStatus, statusValue(model.translator, server), labelWidth, valueWidth),
		"",
		model.detailsRow(i18n.MsgDetailsLastUsed, lastUsedValue(model.translator, server), labelWidth, valueWidth),
		model.detailsRow(i18n.MsgDetailsUseCount, strconv.Itoa(server.UseCount), labelWidth, valueWidth),
	}
	return panelFrameStyle.Width(detailsPanelWidth - 4).Render(strings.Join(lines, "\n"))
}

func (model Model) detailsRow(labelKey, value string, labelWidth, valueWidth int) string {
	label := util.PadRight(model.translator.T(labelKey), labelWidth)
	value = util.Truncate(util.Sanitize(value), valueWidth)
	return panelLabelStyle.Render(label) + panelValueStyle.Render(value)
}

func (model Model) compactDetailsView(server domain.Server) string {
	parts := []string{
		model.translator.T(i18n.MsgDetailsHost) + " " + optionalValue(model.translator, server.Host),
		model.translator.T(i18n.MsgDetailsUser) + " " + optionalValue(model.translator, server.User),
		model.translator.T(i18n.MsgDetailsPort) + " " + strconv.Itoa(server.Port),
		model.translator.T(i18n.MsgDetailsUseCount) + " " + strconv.Itoa(server.UseCount),
	}
	text := util.Truncate(util.Sanitize(strings.Join(parts, " · ")), util.Max(0, model.width-2))
	return "  " + panelValueStyle.Render(text)
}

func optionalValue(translator i18n.Translator, value string) string {
	if strings.TrimSpace(value) == "" {
		return translator.T(i18n.MsgDetailsNone)
	}
	return value
}

func proxyValue(translator i18n.Translator, server domain.Server) string {
	if server.HasProxy {
		return translator.T(i18n.MsgStatusProxy)
	}
	return translator.T(i18n.MsgDetailsNone)
}

func statusValue(translator i18n.Translator, server domain.Server) string {
	if server.HasProxy {
		return translator.T(i18n.MsgStatusProxy)
	}
	if !server.Checked {
		return translator.T(i18n.MsgStatusUnknown)
	}
	if server.Online {
		return translator.T(i18n.MsgStatusOnline)
	}
	return translator.T(i18n.MsgStatusOffline)
}

func lastUsedValue(translator i18n.Translator, server domain.Server) string {
	if server.LastUsed.IsZero() {
		return translator.T(i18n.MsgDetailsNever)
	}
	return server.LastUsed.Format("2006-01-02")
}

func keycap(symbol, label string) string {
	return keycapStyle.Render(symbol) + " " + keyLabelStyle.Render(label)
}

func (model Model) helpBindings() [][2]string {
	return [][2]string{
		{"^n", model.translator.T(i18n.MsgHelpAdd)},
		{"^e", model.translator.T(i18n.MsgHelpEdit)},
		{"^d", model.translator.T(i18n.MsgHelpDelete)},
		{"^c", model.translator.T(i18n.MsgHelpExit)},
	}
}

func (model Model) footerView() string {
	tabLabel := model.translator.T(i18n.MsgHelpMore)
	if model.showHelp {
		tabLabel = model.translator.T(i18n.MsgHelpCollapse)
	}
	primary := "  " + strings.Join([]string{
		keycap("↑↓", model.translator.T(i18n.MsgHelpSelect)),
		keycap("↵", model.translator.T(i18n.MsgHelpConnect)),
		keycap("esc", model.translator.T(i18n.MsgHelpClear)),
		keycap("tab", tabLabel),
	}, "   ")
	if !model.showHelp {
		return primary
	}

	lines := []string{primary, ""}
	for _, binding := range model.helpBindings() {
		lines = append(lines, "  "+keycap(binding[0], binding[1]))
	}
	return strings.Join(lines, "\n")
}

func (model Model) formView() string {
	title := model.translator.T(i18n.MsgFormAddTitle)
	if model.editing {
		title = model.translator.T(i18n.MsgFormEditTitle, util.Sanitize(model.editAlias))
	}
	lines := []string{"", "  " + accentStyle.Render(title), ""}
	labels := []string{
		model.translator.T(i18n.MsgFormName),
		model.translator.T(i18n.MsgFormHost),
		model.translator.T(i18n.MsgFormUser),
		model.translator.T(i18n.MsgFormPort),
		model.translator.T(i18n.MsgFormIdentity),
	}
	for index := range model.form {
		label := metaStyle.Render(util.PadRight(labels[index], 14))
		if index == model.formIndex {
			label = accentStyle.Render(util.PadRight(labels[index], 14))
		}
		lines = append(lines, "  "+label+model.form[index].View())
	}
	lines = append(lines, "")
	if model.errorText != "" {
		lines = append(lines, "  "+accentStyle.Render(model.errorText), "")
	}
	footer := "  " + strings.Join([]string{
		keycap("↑↓", model.translator.T(i18n.MsgHelpField)),
		keycap("↵", model.translator.T(i18n.MsgHelpNext)),
		keycap("^s", model.translator.T(i18n.MsgHelpSave)),
		keycap("esc", model.translator.T(i18n.MsgHelpCancel)),
	}, "   ")
	lines = append(lines, footer)
	return strings.Join(lines, "\n")
}

func (model Model) confirmView() string {
	return strings.Join([]string{
		"",
		"  " + accentStyle.Render(model.translator.T(i18n.MsgDeleteTitle)),
		"",
		"  " + metaStyle.Render(util.Sanitize(model.confirmFor)) + dimStyle.Render(model.translator.T(i18n.MsgDeleteFrom, model.configPath)),
		"",
		"  " + strings.Join([]string{
			keycap("y", model.translator.T(i18n.MsgDeleteAction)),
			keycap("esc", model.translator.T(i18n.MsgHelpCancel)),
		}, "   "),
	}, "\n")
}

func (model Model) connectingView() string {
	alias := model.translator.T(i18n.MsgConnectingFallback)
	if server, ok := model.selectedServer(); ok {
		alias = util.Sanitize(server.Alias)
	}
	return strings.Join([]string{
		"",
		"  " + model.spinner.View() + "  " + boldStyle.Render(alias),
		"",
		dimStyle.Render(model.translator.T(i18n.MsgConnectingHint)),
	}, "\n")
}
