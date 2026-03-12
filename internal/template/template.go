package template

import (
	"errors"
	"strings"
)

const Placeholder = "{{prompt}}"

func Render(layout, prompt string) (string, error) {
	if err := Validate(layout); err != nil {
		return "", err
	}

	return strings.ReplaceAll(layout, Placeholder, prompt), nil
}

func Validate(layout string) error {
	if strings.TrimSpace(layout) == "" {
		return errors.New("template is empty")
	}
	if !strings.Contains(layout, Placeholder) {
		return errors.New("template must contain {{prompt}}")
	}
	return nil
}
