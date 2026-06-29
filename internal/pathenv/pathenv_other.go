//go:build !windows

package pathenv

import (
	"fmt"

	"github.com/elev1e1nSure/hop/internal/apperr"
)

func Apply(action Action) (Result, error) {
	cause := fmt.Errorf("PATH management is only implemented on Windows")
	return Result{}, apperr.Wrap(apperr.ErrPathUnsupported, cause)
}
