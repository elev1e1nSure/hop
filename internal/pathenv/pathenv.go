package pathenv

type Action string

const (
	ActionAdd    Action = "add"
	ActionRemove Action = "remove"
)

type Result struct {
	Directory string
	Changed   bool
}
