package error_handling

import "errors"

type ErrorList []error

// Single will return the error string value
func (e ErrorList) Single() error {
	if e == nil {
		return nil
	}

	switch len(e) {
	case 0:
		return nil

	case 1:
		return e[0]
	}

	var bs []byte
	for _, err := range e {
		if err == nil {
			continue
		}
		bs = append(bs, []byte(err.Error())...)
		bs = append(bs, ',', '\n')
	}

	return errors.New(string(bs))
}
