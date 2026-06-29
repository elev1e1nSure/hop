package i18n

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"sshm/internal/apperr"
)

type Language string

const (
	English Language = "en"
	Russian Language = "ru"
)

const (
	MsgUsage               = "usage"
	MsgFilterPlaceholder   = "filter.placeholder"
	MsgHelpAdd             = "help.add"
	MsgHelpEdit            = "help.edit"
	MsgHelpDelete          = "help.delete"
	MsgHelpExit            = "help.exit"
	MsgHelpSelect          = "help.select"
	MsgHelpConnect         = "help.connect"
	MsgHelpClear           = "help.clear"
	MsgHelpMore            = "help.more"
	MsgHelpCollapse        = "help.collapse"
	MsgHelpField           = "help.field"
	MsgHelpNext            = "help.next"
	MsgHelpSave            = "help.save"
	MsgHelpCancel          = "help.cancel"
	MsgFormAddTitle        = "form.add_title"
	MsgFormEditTitle       = "form.edit_title"
	MsgFormName            = "form.name"
	MsgFormHost            = "form.host"
	MsgFormUser            = "form.user"
	MsgFormPort            = "form.port"
	MsgFormIdentity        = "form.identity"
	MsgPlaceholderName     = "placeholder.name"
	MsgPlaceholderHost     = "placeholder.host"
	MsgPlaceholderUser     = "placeholder.user"
	MsgPlaceholderPort     = "placeholder.port"
	MsgPlaceholderIdentity = "placeholder.identity"
	MsgDeleteTitle         = "delete.title"
	MsgDeleteFrom          = "delete.from"
	MsgDeleteAction        = "delete.action"
	MsgConnectingFallback  = "connecting.fallback"
	MsgConnectingHint      = "connecting.hint"
	MsgValidationAlias     = "validation.alias"
	MsgValidationHost      = "validation.host"
	MsgValidationPort      = "validation.port"
	MsgValidationDuplicate = "validation.duplicate"
)

type Translator struct {
	language Language
}

func New(language Language) Translator {
	if language != Russian {
		language = English
	}
	return Translator{language: language}
}

func (t Translator) Language() Language { return t.language }

func ParseLanguage(value string) (Language, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "en":
		return English, true
	case "ru":
		return Russian, true
	default:
		return "", false
	}
}

func Detect(getenv func(string) string) Language {
	if getenv == nil {
		getenv = os.Getenv
	}
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		value := strings.ToLower(strings.TrimSpace(getenv(name)))
		if strings.HasPrefix(value, "ru") {
			return Russian
		}
		if strings.HasPrefix(value, "en") {
			return English
		}
	}
	return English
}

func (t Translator) T(key string, args ...any) string {
	catalog := english
	if t.language == Russian {
		catalog = russian
	}
	format, ok := catalog[key]
	if !ok {
		format = english[key]
	}
	if format == "" {
		format = key
	}
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}

func (t Translator) Error(err error) string {
	if err == nil {
		return ""
	}
	var localized *apperr.Error
	if errors.Is(err, os.ErrPermission) {
		if errors.As(err, &localized) {
			for _, argument := range localized.Args {
				if path, ok := argument.(string); ok && path != "" {
					return t.T(apperr.ErrPermissionDenied, path)
				}
			}
		}
		return t.T(apperr.ErrPermissionDeniedAny)
	}
	if errors.As(err, &localized) {
		return t.T(localized.Key, localized.Args...)
	}
	return err.Error()
}

var english = map[string]string{
	MsgUsage:               "Usage: sshm [--language en|ru]\n\nOptions:\n  --language en|ru  Interface language; defaults to the current locale\n  -h, --help        Show this help",
	MsgFilterPlaceholder:   "start typing…",
	MsgHelpAdd:             "add",
	MsgHelpEdit:            "edit",
	MsgHelpDelete:          "delete",
	MsgHelpExit:            "quit",
	MsgHelpSelect:          "select",
	MsgHelpConnect:         "connect",
	MsgHelpClear:           "clear",
	MsgHelpMore:            "more",
	MsgHelpCollapse:        "collapse",
	MsgHelpField:           "field",
	MsgHelpNext:            "next",
	MsgHelpSave:            "save",
	MsgHelpCancel:          "cancel",
	MsgFormAddTitle:        "Add server",
	MsgFormEditTitle:       "Edit · %s",
	MsgFormName:            "Name",
	MsgFormHost:            "Host",
	MsgFormUser:            "User",
	MsgFormPort:            "Port",
	MsgFormIdentity:        "Identity file",
	MsgPlaceholderName:     "prod-web-01",
	MsgPlaceholderHost:     "10.0.0.10 or host.example.com",
	MsgPlaceholderUser:     "deploy",
	MsgPlaceholderPort:     "22",
	MsgPlaceholderIdentity: "~/.ssh/id_ed25519",
	MsgDeleteTitle:         "Delete server",
	MsgDeleteFrom:          " from %s",
	MsgDeleteAction:        "delete",
	MsgConnectingFallback:  "server",
	MsgConnectingHint:      "  The TUI will close before ssh starts",
	MsgValidationAlias:     "Name is required and cannot contain whitespace or wildcard characters.",
	MsgValidationHost:      "Host is required.",
	MsgValidationPort:      "Port must be an integer from 1 to 65535.",
	MsgValidationDuplicate: "A server with that name already exists.",

	apperr.ErrHomeDir:             "Could not determine the user home directory.",
	apperr.ErrPermissionDenied:    "Permission denied for %q.",
	apperr.ErrPermissionDeniedAny: "Permission denied for the requested file or directory operation.",
	apperr.ErrInvalidLanguage:     "Unsupported interface language %q. Use en or ru.",
	apperr.ErrMissingLanguage:     "The --language flag requires en or ru.",
	apperr.ErrUnknownArgument:     "Unknown argument %q.",
	apperr.ErrCreateDirectory:     "Could not create directory %q.",
	apperr.ErrReadSSHConfig:       "Could not read SSH config %q.",
	apperr.ErrStatSSHConfig:       "Could not inspect SSH config %q.",
	apperr.ErrUnclosedQuote:       "SSH config %q is malformed: unclosed quote on line %d.",
	apperr.ErrMissingDirectiveArg: "SSH config %q is malformed: directive %q has no value on line %d.",
	apperr.ErrInvalidHostPattern:  "SSH config %q is malformed: invalid Host pattern %q on line %d.",
	apperr.ErrInvalidPortConfig:   "SSH config %q contains invalid port %q on line %d.",
	apperr.ErrWriteSSHConfig:      "Could not write SSH config %q.",
	apperr.ErrServerNotFound:      "Server %q was not found in %s.",
	apperr.ErrDetachServer:        "Server %q could not be detached from its Host block.",
	apperr.ErrAliasNotInHost:      "Server %q was not found in its Host line.",
	apperr.ErrReadHistory:         "Could not read connection history %q.",
	apperr.ErrParseHistory:        "Connection history %q is malformed.",
	apperr.ErrMarshalHistory:      "Could not encode connection history %q.",
	apperr.ErrWriteHistory:        "Could not write connection history %q.",
	apperr.ErrCreateTemporaryFile: "Could not create a temporary file in %q.",
	apperr.ErrSetFileMode:         "Could not set permissions for %q.",
	apperr.ErrWriteTemporaryFile:  "Could not write temporary data for %q.",
	apperr.ErrSyncTemporaryFile:   "Could not flush temporary data for %q.",
	apperr.ErrCloseTemporaryFile:  "Could not close the temporary file for %q.",
	apperr.ErrReplaceFile:         "Could not atomically replace %q.",
	apperr.ErrSSHUnavailable:      "The ssh executable is unavailable or is not present in PATH.",
	apperr.ErrSSHHomeDir:          "Could not expand the identity file path because the home directory is unavailable.",
	apperr.ErrSSHStart:            "Could not start ssh for %s.",
	apperr.ErrSSHExit:             "ssh for %s exited with status %d.",
	apperr.ErrTUI:                 "The terminal interface stopped with an error.",
}

var russian = map[string]string{
	MsgUsage:               "Использование: sshm [--language en|ru]\n\nПараметры:\n  --language en|ru  Язык интерфейса; по умолчанию определяется по локали\n  -h, --help        Показать эту справку",
	MsgFilterPlaceholder:   "начни печатать…",
	MsgHelpAdd:             "добавить",
	MsgHelpEdit:            "изменить",
	MsgHelpDelete:          "удалить",
	MsgHelpExit:            "выход",
	MsgHelpSelect:          "выбор",
	MsgHelpConnect:         "подключиться",
	MsgHelpClear:           "очистить",
	MsgHelpMore:            "ещё",
	MsgHelpCollapse:        "свернуть",
	MsgHelpField:           "поле",
	MsgHelpNext:            "далее",
	MsgHelpSave:            "сохранить",
	MsgHelpCancel:          "отмена",
	MsgFormAddTitle:        "Добавить сервер",
	MsgFormEditTitle:       "Изменить · %s",
	MsgFormName:            "Имя",
	MsgFormHost:            "Хост",
	MsgFormUser:            "Пользователь",
	MsgFormPort:            "Порт",
	MsgFormIdentity:        "Файл ключа",
	MsgPlaceholderName:     "prod-web-01",
	MsgPlaceholderHost:     "10.0.0.10 или host.example.com",
	MsgPlaceholderUser:     "deploy",
	MsgPlaceholderPort:     "22",
	MsgPlaceholderIdentity: "~/.ssh/id_ed25519",
	MsgDeleteTitle:         "Удалить сервер",
	MsgDeleteFrom:          " из %s",
	MsgDeleteAction:        "удалить",
	MsgConnectingFallback:  "сервер",
	MsgConnectingHint:      "  Интерфейс закроется перед запуском ssh",
	MsgValidationAlias:     "Имя обязательно и не должно содержать пробелы или символы шаблонов.",
	MsgValidationHost:      "Хост обязателен.",
	MsgValidationPort:      "Порт должен быть целым числом от 1 до 65535.",
	MsgValidationDuplicate: "Сервер с таким именем уже существует.",

	apperr.ErrHomeDir:             "Не удалось определить домашний каталог пользователя.",
	apperr.ErrPermissionDenied:    "Недостаточно прав для доступа к %q.",
	apperr.ErrPermissionDeniedAny: "Недостаточно прав для операции с файлом или каталогом.",
	apperr.ErrInvalidLanguage:     "Язык интерфейса %q не поддерживается. Используйте en или ru.",
	apperr.ErrMissingLanguage:     "Для флага --language требуется значение en или ru.",
	apperr.ErrUnknownArgument:     "Неизвестный аргумент %q.",
	apperr.ErrCreateDirectory:     "Не удалось создать каталог %q.",
	apperr.ErrReadSSHConfig:       "Не удалось прочитать SSH-конфиг %q.",
	apperr.ErrStatSSHConfig:       "Не удалось получить сведения об SSH-конфиге %q.",
	apperr.ErrUnclosedQuote:       "SSH-конфиг %q повреждён: незакрытая кавычка в строке %d.",
	apperr.ErrMissingDirectiveArg: "SSH-конфиг %q повреждён: у директивы %q нет значения в строке %d.",
	apperr.ErrInvalidHostPattern:  "SSH-конфиг %q повреждён: некорректный шаблон Host %q в строке %d.",
	apperr.ErrInvalidPortConfig:   "В SSH-конфиге %q указан некорректный порт %q в строке %d.",
	apperr.ErrWriteSSHConfig:      "Не удалось записать SSH-конфиг %q.",
	apperr.ErrServerNotFound:      "Сервер %q не найден в %s.",
	apperr.ErrDetachServer:        "Не удалось отделить сервер %q от его блока Host.",
	apperr.ErrAliasNotInHost:      "Сервер %q не найден в строке Host.",
	apperr.ErrReadHistory:         "Не удалось прочитать историю подключений %q.",
	apperr.ErrParseHistory:        "История подключений %q повреждена.",
	apperr.ErrMarshalHistory:      "Не удалось подготовить историю подключений %q к сохранению.",
	apperr.ErrWriteHistory:        "Не удалось записать историю подключений %q.",
	apperr.ErrCreateTemporaryFile: "Не удалось создать временный файл в каталоге %q.",
	apperr.ErrSetFileMode:         "Не удалось установить права доступа для %q.",
	apperr.ErrWriteTemporaryFile:  "Не удалось записать временные данные для %q.",
	apperr.ErrSyncTemporaryFile:   "Не удалось сбросить временные данные на диск для %q.",
	apperr.ErrCloseTemporaryFile:  "Не удалось закрыть временный файл для %q.",
	apperr.ErrReplaceFile:         "Не удалось атомарно заменить файл %q.",
	apperr.ErrSSHUnavailable:      "Исполняемый файл ssh недоступен или отсутствует в PATH.",
	apperr.ErrSSHHomeDir:          "Не удалось раскрыть путь к файлу ключа: домашний каталог недоступен.",
	apperr.ErrSSHStart:            "Не удалось запустить ssh для %s.",
	apperr.ErrSSHExit:             "ssh для %s завершился с кодом %d.",
	apperr.ErrTUI:                 "Терминальный интерфейс завершился с ошибкой.",
}
