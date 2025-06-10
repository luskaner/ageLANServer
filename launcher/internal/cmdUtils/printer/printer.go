package printer

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/launcher-common/executor/exec"
	"github.com/spf13/viper"
	"strconv"
	"strings"
)

const (
	Error         = "❌"
	Warning       = "⚠️"
	Info          = "ℹ️"
	Debug         = "🐛"
	Success       = "✔️"
	Clean         = "🧹"
	Configuration = "⚙️"
	Search        = "🔍"
	Execute       = "▶️"
	AllDone       = "🎉"
	Stop          = "🛑"
	Speed         = "⚡️"
)

type StyledText struct {
	text  string
	style *lipgloss.Style
}

type StyledTextOrText interface {
	*StyledText | string
}

func T(text string) *StyledText {
	return &StyledText{
		text: text,
	}
}

func TS(text string, style lipgloss.Style) *StyledText {
	return &StyledText{
		text:  text,
		style: &style,
	}
}

var (
	textStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250"))

	FilePathStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	ComponentStyle = lipgloss.NewStyle().Bold(true)
	LiteralStyle   = lipgloss.NewStyle().Italic(true)
	OptionStyle    = lipgloss.NewStyle().Italic(true).Bold(true)
	SeparatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func render[T StyledTextOrText](level string, value T) string {
	baseStyle := textStyle
	switch level {
	case Debug:
		baseStyle = baseStyle.Foreground(lipgloss.Color("245"))
	case Error:
		baseStyle = baseStyle.Foreground(lipgloss.Color("9"))
	}
	var text string
	switch val := any(value).(type) {
	case string:
		text = val
	case *StyledText:
		text = val.text
		if val.style != nil {
			baseStyle = val.style.Inherit(baseStyle)
		}
	}
	return baseStyle.Render(text)
}

func generatePrefix(level string, prefix string) string {
	return prefix + level
}

func Gen[T StyledTextOrText](level string, prefix string, styledTexts ...T) string {
	if level == Debug && !viper.GetBool("Config.Debug") {
		return ""
	}
	finalPrefix := generatePrefix(level, prefix)
	var suffixBuilder strings.Builder
	for _, styledText := range styledTexts {
		suffixBuilder.WriteString(render(level, styledText))
	}
	return joinMessage(finalPrefix, suffixBuilder.String())
}

func genSimpLn(level string, text string) string {
	return Gen(level, "", text)
}

func Print[T StyledTextOrText](level string, prefix string, styledTexts ...T) {
	if str := Gen(level, prefix, styledTexts...); str != "" {
		fmt.Print(str)
	}
}

func Println[T StyledTextOrText](level string, styledTexts ...T) {
	if str := Gen(level, "", styledTexts...); str != "" {
		fmt.Println(str)
	}
}

func PrintSimpln(level string, text string) {
	if str := genSimpLn(level, text); str != "" {
		fmt.Println(str)
	}
}

func joinMessage(prefix string, suffix string) string {
	return lipgloss.JoinHorizontal(lipgloss.Left, prefix, " ", suffix)
}

func PrintInvalidOption(option string, value string, validValues ...string) {
	valuesFormatted := []*StyledText{
		T(`Invalid value "`),
		TS(
			value,
			LiteralStyle,
		),
		T(`" for `),
		TS(
			option,
			OptionStyle,
		),
		T(`. Valid values are: `),
	}
	for i, v := range validValues {
		valuesFormatted = append(valuesFormatted, T(`"`))
		valuesFormatted = append(valuesFormatted, TS(v, LiteralStyle))
		valuesFormatted = append(valuesFormatted, T(`"`))
		if i < len(validValues)-2 {
			valuesFormatted = append(valuesFormatted, T(", "))
		} else if i < len(validValues)-1 {
			valuesFormatted = append(valuesFormatted, T(" or "))
		}
	}
	valuesFormatted = append(valuesFormatted, T("."))
	Println(
		Error,
		valuesFormatted...,
	)
}

func PrintFailedParseOption(option string, err error) {
	Println(
		Error,
		T(`Failed to parse `),
		TS(option, OptionStyle),
		T(`.`),
		TS(err.Error(), LiteralStyle),
	)
	PrintSimpln(Debug, err.Error())
}

func PrintResultError(result *exec.Result) {
	if result == nil {
		return
	}
	if result.Err != nil {
		PrintError(result.Err)
	}
	if result.ExitCode != common.ErrSuccess {
		Println(
			Debug,
			T("Exit code: "),
			TS(strconv.Itoa(result.ExitCode), LiteralStyle),
		)
	}
}

func PrintFailedResultError(result *exec.Result) {
	PrintFailed()
	PrintResultError(result)
}

func PrintFailedError(err error) {
	PrintFailed()
	PrintError(err)
}

func PrintError(err error) {
	Println(
		Debug,
		T("Error: "),
		TS(err.Error(), LiteralStyle),
	)
}

func PrintFailed() {
	PrintSimpln(Error, "failed.")
}

func PrintSucceeded() {
	PrintSimpln(Success, "succeeded.")
}
