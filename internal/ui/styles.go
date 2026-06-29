package ui

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/elev1e1nSure/hop/internal/domain"
	"github.com/elev1e1nSure/hop/internal/util"
)

const selectedBackground = lipgloss.Color("#3A2A22")

var (
	accentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E0976A")).Bold(true)
	dimStyle    = lipgloss.NewStyle().Faint(true)
	boldStyle   = lipgloss.NewStyle().Bold(true)

	aliasStyle      = lipgloss.NewStyle().Bold(true)
	metaStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#7A7A82"))
	onlineStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#7FB582"))
	selectedBar     = lipgloss.NewStyle().Background(selectedBackground)
	selectedAlias   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2A878")).Bold(true).Background(selectedBackground)
	selectedMeta    = lipgloss.NewStyle().Foreground(lipgloss.Color("#C9A48E")).Background(selectedBackground)
	keycapStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#DCDCDC")).Background(lipgloss.Color("#2C2C34")).Padding(0, 1)
	keyLabelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#8A8A92"))
	onlineDotStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#7FB582"))
	offlineDotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E06C75"))
	unknownDotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7A7A82"))
	pointerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2A878")).Background(selectedBackground).Bold(true)

	filterBackground = lipgloss.Color("#1E1E1E")
	filterBoxStyle   = lipgloss.NewStyle().Background(filterBackground).PaddingRight(1)

	panelFrameStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#2C2C34")).Padding(0, 1)
	panelTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2A878")).Bold(true)
	panelLabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7A7A82"))
	panelValueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#DCDCDC"))
)

type rowItem struct {
	Server domain.Server
}

func (item rowItem) FilterValue() string {
	return strings.Join([]string{item.Server.Alias, item.Server.Host, item.Server.User, item.Server.IdentityFile}, " ")
}

type rowDelegate struct{}

func (rowDelegate) Height() int                         { return 1 }
func (rowDelegate) Spacing() int                        { return 0 }
func (rowDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (rowDelegate) Render(writer io.Writer, model list.Model, index int, item list.Item) {
	row, ok := item.(rowItem)
	if !ok {
		return
	}
	width := model.Width()
	if width < 4 {
		return
	}

	selected := index == model.Index()
	target := util.Sanitize(row.Server.Host)
	if row.Server.User != "" {
		target = util.Sanitize(row.Server.User) + "@" + target
	}
	if row.Server.Port != 22 {
		target += ":" + strconv.Itoa(row.Server.Port)
	}
	metadata := target

	const padding = 2
	const statusWidth = 2
	alias := util.Sanitize(row.Server.Alias)
	available := width - padding*2 - statusWidth
	if utf8.RuneCountInString(alias)+1+utf8.RuneCountInString(metadata) > available {
		metadataMax := util.Max(0, available-utf8.RuneCountInString(alias)-1)
		metadata = util.Truncate(metadata, metadataMax)
	}
	if utf8.RuneCountInString(alias) > available-1 {
		alias = util.Truncate(alias, util.Max(1, available-1))
	}
	gap := util.Max(1, available-utf8.RuneCountInString(alias)-utf8.RuneCountInString(metadata))

	dotStyle := unknownDotStyle
	dot := "◌"
	if row.Server.Checked {
		if row.Server.Online {
			dot = "●"
			dotStyle = onlineDotStyle
		} else {
			dot = "○"
			dotStyle = offlineDotStyle
		}
	}

	if selected {
		dotStyle = dotStyle.Background(selectedBackground)
		_, _ = fmt.Fprint(writer, pointerStyle.Render("❯ ")+
			dotStyle.Render(dot+" ")+
			selectedAlias.Render(alias)+
			selectedBar.Render(strings.Repeat(" ", gap))+
			selectedMeta.Render(metadata)+
			selectedBar.Render(strings.Repeat(" ", padding)))
		return
	}

	metadataStyle := metaStyle
	if row.Server.Checked && row.Server.Online {
		metadataStyle = onlineStyle
	}
	_, _ = fmt.Fprint(writer, strings.Repeat(" ", padding)+
		dotStyle.Render(dot+" ")+
		aliasStyle.Render(alias)+
		strings.Repeat(" ", gap)+
		metadataStyle.Render(metadata)+
		strings.Repeat(" ", padding))
}
