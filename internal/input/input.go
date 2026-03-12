package input

import (
	"errors"
	"io"
	"os"
	"strings"
)

var ErrNoInput = errors.New("no input provided")

func Resolve(args []string, stdin io.Reader, stdinPiped bool) (string, error) {
	if len(args) > 0 {
		text := strings.TrimSpace(strings.Join(args, " "))
		if text == "" {
			return "", ErrNoInput
		}
		return text, nil
	}

	if !stdinPiped {
		return "", ErrNoInput
	}

	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", err
	}

	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", ErrNoInput
	}

	return text, nil
}

func StdinIsPiped(file *os.File) (bool, error) {
	info, err := file.Stat()
	if err != nil {
		return false, err
	}

	return (info.Mode() & os.ModeCharDevice) == 0, nil
}
