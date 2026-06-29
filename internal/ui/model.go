package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"hop/internal/apperr"
	"hop/internal/domain"
	"hop/internal/history"
	"hop/internal/i18n"
	"hop/internal/sshconfig"
	"hop/internal/util"
)

type appMode int

const (
	modeBrowse appMode = iota
	modeForm
	modeConfirmDelete
	modeConnecting
)

const (
	detailsPanelMinWidth = 96
	detailsPanelWidth    = 32
	detailsPanelGap      = 2
)

type Result struct {
	Connect *domain.Server
	Binary  string
	History map[string]domain.HistoryRecord
}

type Model struct {
	configPath  string
	historyPath string
	translator  i18n.Translator
	config      *sshconfig.Config
	history     map[string]domain.HistoryRecord
	servers     []domain.Server
	list        list.Model
	filter      textinput.Model
	mode        appMode
	showHelp    bool
	width       int
	height      int
	errorText   string

	form       []textinput.Model
	formIndex  int
	editing    bool
	editAlias  string
	confirmFor string

	spinner spinner.Model
	connect *domain.Server
	binary  string
}

func Run(configPath, historyPath string, translator i18n.Translator) (Result, error) {
	model, err := newModel(configPath, historyPath, translator)
	if err != nil {
		return Result{}, err
	}
	program := tea.NewProgram(model, tea.WithAltScreen())
	final, err := program.Run()
	if err != nil {
		return Result{}, apperr.Wrap(apperr.ErrTUI, fmt.Errorf("run terminal UI: %w", err))
	}
	finished, ok := final.(Model)
	if !ok {
		return Result{}, apperr.Wrap(apperr.ErrTUI, fmt.Errorf("unexpected terminal UI model type %T", final))
	}
	return Result{Connect: finished.connect, Binary: finished.binary, History: finished.history}, nil
}

func newModel(configPath, historyPath string, translator i18n.Translator) (Model, error) {
	config, err := sshconfig.Load(configPath)
	if err != nil {
		return Model{}, err
	}
	records, err := history.Load(historyPath)
	if err != nil {
		return Model{}, err
	}
	servers, err := sshconfig.ResolveServers(config, records)
	if err != nil {
		return Model{}, err
	}

	listModel := list.New(nil, rowDelegate{}, 100, 20)
	listModel.SetShowTitle(false)
	listModel.SetShowFilter(false)
	listModel.SetShowStatusBar(false)
	listModel.SetShowPagination(false)
	listModel.SetShowHelp(false)
	listModel.SetFilteringEnabled(false)
	listModel.DisableQuitKeybindings()

	filter := textinput.New()
	filter.Prompt = ""
	filter.Placeholder = translator.T(i18n.MsgFilterPlaceholder)
	filter.CharLimit = 128
	filter.TextStyle = lipgloss.NewStyle().Background(filterBackground)
	filter.PlaceholderStyle = lipgloss.NewStyle().Faint(true).Background(filterBackground)
	filter.Cursor.Style = lipgloss.NewStyle().Background(filterBackground).Reverse(true)
	filter.Focus()

	progress := spinner.New()
	progress.Spinner = spinner.Dot
	progress.Style = accentStyle

	model := Model{
		configPath:  configPath,
		historyPath: historyPath,
		translator:  translator,
		config:      config,
		history:     records,
		servers:     servers,
		list:        listModel,
		filter:      filter,
		mode:        modeBrowse,
		width:       100,
		height:      28,
		spinner:     progress,
	}
	model.resizeList()
	model.refreshItems("")
	return model, nil
}

func (model Model) Init() tea.Cmd {
	return tea.Batch(model.checkAllCmd(), textinput.Blink)
}

func (model Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	var command tea.Cmd

	switch message := message.(type) {
	case tea.WindowSizeMsg:
		model.width = util.Max(30, message.Width)
		model.height = util.Max(10, message.Height)
		model.resizeList()
		model.resizeForm()
		return model, nil
	case statusCheckMsg:
		selected := model.selectedAlias()
		for index := range model.servers {
			if model.servers[index].Alias == message.Alias {
				model.servers[index].Checked = true
				model.servers[index].Online = message.Online
				break
			}
		}
		model.refreshItems(selected)
		return model, nil
	case connectReadyMsg:
		if message.Err != nil {
			model.mode = modeBrowse
			model.errorText = model.translator.Error(message.Err)
			model.resizeList()
			return model, nil
		}
		for index := range model.servers {
			if model.servers[index].Alias == message.Server.Alias {
				model.servers[index].Checked = true
				model.servers[index].Online = message.Online
				break
			}
		}
		server := message.Server
		model.connect = &server
		model.binary = message.Binary
		return model, tea.Quit
	case spinner.TickMsg:
		if model.mode == modeConnecting {
			model.spinner, command = model.spinner.Update(message)
			return model, command
		}
	}

	if key, ok := message.(tea.KeyMsg); ok && key.String() == "ctrl+c" {
		return model, tea.Quit
	}

	switch model.mode {
	case modeForm:
		return model.updateForm(message)
	case modeConfirmDelete:
		return model.updateDeleteConfirm(message)
	case modeConnecting:
		return model, nil
	default:
		return model.updateBrowse(message)
	}
}

func (model Model) updateBrowse(message tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := message.(tea.KeyMsg)
	if !ok {
		return model, nil
	}

	switch key.String() {
	case "up", "ctrl+k":
		model.list.CursorUp()
		return model, nil
	case "down", "ctrl+j":
		model.list.CursorDown()
		return model, nil
	case "esc":
		model.filter.SetValue("")
		model.errorText = ""
		model.refreshItems("")
		model.resizeList()
		return model, nil
	case "enter":
		if server, ok := model.selectedServer(); ok {
			model.mode = modeConnecting
			model.errorText = ""
			return model, tea.Batch(model.spinner.Tick, prepareConnectCmd(server))
		}
		return model, nil
	case "ctrl+n":
		model.startAdd()
		return model, textinput.Blink
	case "ctrl+e":
		if server, ok := model.selectedServer(); ok {
			model.startEdit(server)
			return model, textinput.Blink
		}
		return model, nil
	case "ctrl+d":
		if server, ok := model.selectedServer(); ok {
			model.confirmFor = server.Alias
			model.mode = modeConfirmDelete
			model.errorText = ""
		}
		return model, nil
	case "tab":
		model.showHelp = !model.showHelp
		model.resizeList()
		return model, nil
	}

	before := model.filter.Value()
	var command tea.Cmd
	model.filter, command = model.filter.Update(message)
	if model.filter.Value() != before {
		model.refreshItems("")
	}
	return model, command
}

func (model Model) updateDeleteConfirm(message tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := message.(tea.KeyMsg)
	if !ok {
		return model, nil
	}
	switch strings.ToLower(key.String()) {
	case "n", "esc":
		model.mode = modeBrowse
		model.confirmFor = ""
		return model, nil
	case "y", "enter":
		alias := model.confirmFor
		if err := sshconfig.DeleteServer(model.config, alias); err != nil {
			model.errorText = model.translator.Error(err)
			model.mode = modeBrowse
			model.resizeList()
			return model, nil
		}
		delete(model.history, alias)
		historyErr := history.Save(model.historyPath, model.history)
		reloadErr := model.reloadConfig("")
		model.mode = modeBrowse
		model.confirmFor = ""
		switch {
		case historyErr != nil:
			model.errorText = model.translator.Error(historyErr)
		case reloadErr != nil:
			model.errorText = model.translator.Error(reloadErr)
		default:
			model.errorText = ""
		}
		model.resizeList()
		return model, model.checkAllCmd()
	}
	return model, nil
}

func (model *Model) resizeList() {
	extra := 13
	if model.showHelp {
		extra += 5
	}
	if !model.showDetailsPanel() {
		extra++
	}
	if model.errorText != "" {
		extra += 2
	}
	model.list.SetSize(model.listWidth(), util.Max(3, model.height-extra))
}

func (model Model) showDetailsPanel() bool {
	return model.width >= detailsPanelMinWidth
}

func (model Model) listWidth() int {
	if !model.showDetailsPanel() {
		return model.width
	}
	available := model.width - detailsPanelWidth - detailsPanelGap
	return util.Max(30, available)
}

func (model *Model) resizeForm() {
	for index := range model.form {
		model.form[index].Width = util.Max(10, model.width-18)
	}
}

func (model *Model) refreshItems(preferredAlias string) {
	if preferredAlias == "" {
		preferredAlias = model.selectedAlias()
	}
	query := strings.TrimSpace(strings.ToLower(model.filter.Value()))
	filtered := make([]domain.Server, 0, len(model.servers))
	for _, server := range model.servers {
		target := strings.ToLower(strings.Join([]string{server.Alias, server.Host, server.User, server.IdentityFile}, " "))
		if util.FuzzyMatch(query, target) {
			filtered = append(filtered, server)
		}
	}

	items := make([]list.Item, 0, len(filtered))
	for _, server := range filtered {
		items = append(items, rowItem{Server: server})
	}
	model.list.SetItems(items)

	selected := -1
	for index, item := range items {
		row := item.(rowItem)
		if row.Server.Alias == preferredAlias {
			selected = index
			break
		}
	}
	if selected < 0 && len(items) > 0 {
		selected = 0
	}
	if selected >= 0 {
		model.list.Select(selected)
	} else {
		model.list.Select(0)
	}
}

func (model Model) selectedAlias() string {
	if server, ok := model.selectedServer(); ok {
		return server.Alias
	}
	return ""
}

func (model Model) selectedServer() (domain.Server, bool) {
	item := model.list.SelectedItem()
	if item == nil {
		return domain.Server{}, false
	}
	row, ok := item.(rowItem)
	if !ok {
		return domain.Server{}, false
	}
	return row.Server, true
}

func (model *Model) reloadConfig(preferredAlias string) error {
	statuses := make(map[string][2]bool, len(model.servers))
	for _, server := range model.servers {
		statuses[server.Alias] = [2]bool{server.Checked, server.Online}
	}
	config, err := sshconfig.Load(model.configPath)
	if err != nil {
		return err
	}
	servers, err := sshconfig.ResolveServers(config, model.history)
	if err != nil {
		return err
	}
	model.config = config
	model.servers = servers
	for index := range model.servers {
		if status, ok := statuses[model.servers[index].Alias]; ok {
			model.servers[index].Checked = status[0]
			model.servers[index].Online = status[1]
		}
	}
	model.refreshItems(preferredAlias)
	return nil
}
