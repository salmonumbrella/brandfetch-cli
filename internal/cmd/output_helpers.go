package cmd

import (
	"io"
	"os"

	"github.com/salmonumbrella/brandfetch-cli/internal/output"
	"golang.org/x/term"
)

func resolveOutput(cmd outWriterProvider) (output.Format, bool, error) {
	formatInput := outputFormat
	if formatInput == "" {
		formatInput = "text"
	}
	format, err := output.ParseFormat(formatInput)
	if err != nil {
		return format, false, err
	}

	modeInput := colorMode
	if modeInput == "" {
		modeInput = "auto"
	}
	mode, err := output.ParseColorMode(modeInput)
	if err != nil {
		return format, false, err
	}

	noColor := os.Getenv("NO_COLOR") != ""
	isTTY := isTerminal(cmd.OutOrStdout())
	colorize := output.ResolveColorMode(mode, format, noColor, isTTY)
	return format, colorize, nil
}

type outWriterProvider interface {
	OutOrStdout() io.Writer
}

func isTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(file.Fd()))
}
