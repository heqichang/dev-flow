package ui

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
)

var (
	InfoColor    = color.New(color.FgCyan).SprintFunc()
	SuccessColor = color.New(color.FgGreen, color.Bold).SprintFunc()
	ErrorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	WarningColor = color.New(color.FgYellow).SprintFunc()
	HeaderColor  = color.New(color.FgMagenta, color.Bold).SprintFunc()
)

func Info(msg string) {
	fmt.Println(InfoColor("ℹ ") + msg)
}

func Success(msg string) {
	fmt.Println(SuccessColor("✓ ") + msg)
}

func Error(msg string) {
	fmt.Println(ErrorColor("✗ ") + msg)
}

func Warning(msg string) {
	fmt.Println(WarningColor("⚠ ") + msg)
}

func Header(msg string) {
	fmt.Println(HeaderColor("\n" + msg))
	fmt.Println(HeaderColor(strings.Repeat("=", len(msg)))
}

func Title(msg string) {
	fmt.Println(HeaderColor("\n" + msg))
}

func Step(index int, total int, msg string) {
	fmt.Printf("%s [%d/%d] %s\n", InfoColor("→"), index, total, msg)
}

func Spinner(msg string) *SpinnerWrapper {
	return NewSpinner(msg)
}

func ProgressBar(total int, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
}

func Prompt(msg string) {
	fmt.Print(InfoColor("? ") + msg + " ")
}
