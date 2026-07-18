package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ANSI Escape Codes
const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Black     = "\033[30m"
	Red       = "\033[31m"
	Green     = "\033[32m"
	Yellow    = "\033[33m"
	Blue      = "\033[34m"
	Magenta   = "\033[35m"
	Cyan      = "\033[36m"
	White     = "\033[37m"
	Gray      = "\033[90m"
	LightRed  = "\033[91m"
	LightGreen= "\033[92m"
	LightCyan = "\033[96m"
)

// Symbols
const (
	Check = "✔"
	Cross = "✘"
	Warn  = "▲"
	Info  = "○"
	Bullet= "◇"
)

func Header(title string) {
	fmt.Printf("\n%s%s┌  %s%s\n", Bold, LightCyan, title, Reset)
	fmt.Printf("%s%s│%s\n", Bold, LightCyan, Reset)
}

func Footer() {
	fmt.Printf("%s%s│%s\n", Bold, LightCyan, Reset)
	fmt.Printf("%s%s└%s\n\n", Bold, LightCyan, Reset)
}

func Step(status, msg string) {
	fmt.Printf("%s%s│  %s %s%s\n", Bold, LightCyan, status, msg, Reset)
}

func BulletPoint(label, value string) {
	fmt.Printf("%s%s│  %s %s%s%s: %s%s\n", Bold, LightCyan, Gray, Bullet, Reset, Bold, label, Reset, value)
}

func LogSuccess(msg string) {
	fmt.Printf("%s%s│  %s%s %s%s\n", Bold, LightCyan, Green, Check, msg, Reset)
}

func LogError(msg string) {
	fmt.Printf("%s%s│  %s%s %s%s\n", Bold, LightCyan, Red, Cross, msg, Reset)
}

func LogWarning(msg string) {
	fmt.Printf("%s%s│  %s%s %s%s\n", Bold, LightCyan, Yellow, Warn, msg, Reset)
}

func LogInfo(msg string) {
	fmt.Printf("%s%s│  %s%s %s%s\n", Bold, LightCyan, LightCyan, Info, msg, Reset)
}

func Prompt(label string, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	if defaultValue != "" {
		fmt.Printf("%s%s│  %s%s %s %s(%s): %s", Bold, LightCyan, Gray, Bullet, Reset, label, Gray, defaultValue, Reset)
	} else {
		fmt.Printf("%s%s│  %s%s %s: %s", Bold, LightCyan, Gray, Bullet, Reset, label, Reset)
	}
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func PromptRequired(label string) string {
	for {
		val := Prompt(label, "")
		if val != "" {
			return val
		}
		LogWarning("Input tidak boleh kosong!")
	}
}

func PromptConfirm(label string, defaultYes bool) bool {
	choices := "Y/n"
	if !defaultYes {
		choices = "y/N"
	}
	val := Prompt(fmt.Sprintf("%s [%s]", label, choices), "")
	val = strings.ToLower(strings.TrimSpace(val))
	if val == "" {
		return defaultYes
	}
	return val == "y" || val == "yes"
}
