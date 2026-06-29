package apperr

import "fmt"

const (
	ErrHomeDir             = "error.home_dir"
	ErrPermissionDenied    = "error.permission_denied"
	ErrPermissionDeniedAny = "error.permission_denied_generic"
	ErrInvalidLanguage     = "error.invalid_language"
	ErrMissingLanguage     = "error.missing_language"
	ErrUnknownArgument     = "error.unknown_argument"
	ErrCreateDirectory     = "error.create_directory"
	ErrReadSSHConfig       = "error.read_ssh_config"
	ErrStatSSHConfig       = "error.stat_ssh_config"
	ErrUnclosedQuote       = "error.unclosed_quote"
	ErrMissingDirectiveArg = "error.missing_directive_argument"
	ErrInvalidHostPattern  = "error.invalid_host_pattern"
	ErrInvalidPortConfig   = "error.invalid_port_config"
	ErrWriteSSHConfig      = "error.write_ssh_config"
	ErrServerNotFound      = "error.server_not_found"
	ErrDetachServer        = "error.detach_server"
	ErrAliasNotInHost      = "error.alias_not_in_host"
	ErrReadHistory         = "error.read_history"
	ErrParseHistory        = "error.parse_history"
	ErrMarshalHistory      = "error.marshal_history"
	ErrWriteHistory        = "error.write_history"
	ErrCreateTemporaryFile = "error.create_temporary_file"
	ErrSetFileMode         = "error.set_file_mode"
	ErrWriteTemporaryFile  = "error.write_temporary_file"
	ErrSyncTemporaryFile   = "error.sync_temporary_file"
	ErrCloseTemporaryFile  = "error.close_temporary_file"
	ErrReplaceFile         = "error.replace_file"
	ErrSSHUnavailable      = "error.ssh_unavailable"
	ErrSSHHomeDir          = "error.ssh_home_dir"
	ErrSSHStart            = "error.ssh_start"
	ErrSSHExit             = "error.ssh_exit"
	ErrTUI                 = "error.tui"
)

type Error struct {
	Key   string
	Args  []any
	Cause error
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Key
	}
	return fmt.Sprintf("%s: %v", e.Key, e.Cause)
}

func (e *Error) Unwrap() error { return e.Cause }

func New(key string, args ...any) error {
	return &Error{Key: key, Args: args}
}

func Wrap(key string, cause error, args ...any) error {
	if cause == nil {
		return nil
	}
	return &Error{Key: key, Args: args, Cause: cause}
}
