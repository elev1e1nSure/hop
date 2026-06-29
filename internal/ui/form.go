package ui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/elev1e1nSure/hop/internal/domain"
	"github.com/elev1e1nSure/hop/internal/history"
	"github.com/elev1e1nSure/hop/internal/i18n"
	"github.com/elev1e1nSure/hop/internal/sshconfig"
	"github.com/elev1e1nSure/hop/internal/util"
)

func (model Model) updateForm(message tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := message.(tea.KeyMsg)
	if ok {
		switch key.String() {
		case "esc":
			model.mode = modeBrowse
			model.form = nil
			model.errorText = ""
			return model, nil
		case "tab", "down":
			model.setFormFocus((model.formIndex + 1) % len(model.form))
			return model, textinput.Blink
		case "shift+tab", "up":
			model.setFormFocus((model.formIndex - 1 + len(model.form)) % len(model.form))
			return model, textinput.Blink
		case "enter":
			if model.formIndex < len(model.form)-1 {
				model.setFormFocus(model.formIndex + 1)
				return model, textinput.Blink
			}
			return model.submitForm()
		case "ctrl+s":
			return model.submitForm()
		}
	}

	var command tea.Cmd
	model.form[model.formIndex], command = model.form[model.formIndex].Update(message)
	return model, command
}

func (model *Model) startAdd() {
	model.mode = modeForm
	model.editing = false
	model.editAlias = ""
	model.errorText = ""
	model.form = makeFormInputs(model.width, model.translator)
	model.form[3].SetValue("22")
	model.setFormFocus(0)
}

func (model *Model) startEdit(server domain.Server) {
	model.mode = modeForm
	model.editing = true
	model.editAlias = server.Alias
	model.errorText = ""
	model.form = makeFormInputs(model.width, model.translator)
	model.form[0].SetValue(server.Alias)
	model.form[1].SetValue(server.Host)
	model.form[2].SetValue(server.User)
	model.form[3].SetValue(strconv.Itoa(server.Port))
	model.form[4].SetValue(server.IdentityFile)
	model.setFormFocus(0)
}

func makeFormInputs(width int, translator i18n.Translator) []textinput.Model {
	inputs := make([]textinput.Model, 5)
	placeholders := []string{
		translator.T(i18n.MsgPlaceholderName),
		translator.T(i18n.MsgPlaceholderHost),
		translator.T(i18n.MsgPlaceholderUser),
		translator.T(i18n.MsgPlaceholderPort),
		translator.T(i18n.MsgPlaceholderIdentity),
	}
	for index := range inputs {
		inputs[index] = textinput.New()
		inputs[index].Prompt = ""
		inputs[index].Placeholder = placeholders[index]
		inputs[index].Width = util.Max(10, width-18)
		inputs[index].CharLimit = 512
		inputs[index].Cursor.Style = lipgloss.NewStyle()
	}
	inputs[0].CharLimit = 128
	inputs[2].CharLimit = 128
	inputs[3].CharLimit = 5
	return inputs
}

func (model *Model) setFormFocus(index int) {
	model.formIndex = index
	for current := range model.form {
		if current == index {
			model.form[current].Focus()
		} else {
			model.form[current].Blur()
		}
	}
}

func (model Model) submitForm() (tea.Model, tea.Cmd) {
	server, formErr := model.readFormServer()
	if formErr != "" {
		model.errorText = formErr
		return model, nil
	}

	writeErr := model.applyFormServer(server)
	if writeErr != nil {
		model.errorText = model.translator.Error(writeErr)
		return model, nil
	}

	var historyErr error
	if model.editing && server.Alias != model.editAlias {
		if record, ok := model.history[model.editAlias]; ok {
			model.history[server.Alias] = record
			delete(model.history, model.editAlias)
			historyErr = history.Save(model.historyPath, model.history)
		}
	}

	reloadErr := model.reloadConfig(server.Alias)
	model.closeForm()
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

func (model *Model) closeForm() {
	model.mode = modeBrowse
	model.form = nil
}

func (model Model) readFormServer() (domain.Server, string) {
	alias := strings.TrimSpace(model.form[0].Value())
	host := strings.TrimSpace(model.form[1].Value())
	user := strings.TrimSpace(model.form[2].Value())
	portText := strings.TrimSpace(model.form[3].Value())
	identity := strings.TrimSpace(model.form[4].Value())

	if alias == "" || strings.ContainsAny(alias, " \t\r\n*?[]!") {
		return domain.Server{}, model.translator.T(i18n.MsgValidationAlias)
	}
	if host == "" {
		return domain.Server{}, model.translator.T(i18n.MsgValidationHost)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return domain.Server{}, model.translator.T(i18n.MsgValidationPort)
	}
	for _, server := range model.servers {
		if server.Alias == alias && (!model.editing || alias != model.editAlias) {
			return domain.Server{}, model.translator.T(i18n.MsgValidationDuplicate)
		}
	}
	return domain.Server{Alias: alias, Host: host, User: user, Port: port, IdentityFile: identity}, ""
}

func (model *Model) applyFormServer(server domain.Server) error {
	if model.editing {
		return sshconfig.EditServer(model.config, model.editAlias, server)
	}
	return sshconfig.AddServer(model.config, server)
}
