package output

import "errors"

type printedError struct {
	err error
}

func (e printedError) Error() string {
	return e.err.Error()
}

func (e printedError) Unwrap() error {
	return e.err
}

func PrintedError(err error) error {
	if err == nil {
		return nil
	}
	return printedError{err: err}
}

func IsPrintedError(err error) bool {
	var printed printedError
	return errors.As(err, &printed)
}
