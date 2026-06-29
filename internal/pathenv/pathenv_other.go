//go:build !windows

package pathenv

import (
	"fmt"

	"hop/internal/apperr"
)

func Apply(action Action) (Result, error) {
	cause := fmt.Errorf("PATH management is only implemented on Windows")
	return Result{}, apperr.Wrap(apperr.ErrPathUnsupported, cause)
}
