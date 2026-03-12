package clipboard

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type Writer interface {
	Copy(ctx context.Context, text string) error
}

type PBClipboard struct {
	CommandName string
}

func NewPBClipboard() *PBClipboard {
	return &PBClipboard{CommandName: "pbcopy"}
}

func (c *PBClipboard) Copy(ctx context.Context, text string) error {
	cmd := exec.CommandContext(ctx, c.CommandName)
	cmd.Stdin = bytes.NewBufferString(text)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("copy to clipboard: %w: %s", err, string(output))
	}

	return nil
}
