package regularity

import "errors"

var (
	ErrArgCount = errors.New("arg count mismatch (must be 2)")
	ErrZero = errors.New("count must be > 0")
	ErrFirstArg = errors.New("first argument must be nubmer")
)
