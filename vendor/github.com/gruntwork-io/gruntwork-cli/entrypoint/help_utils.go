package entrypoint

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/tabwriter"
	"text/template"
	"unicode"
)

// Override this to have custom line widths
var HelpTextLineWidth = 80

// This version of the app help template has the following changes over the default ones:
// - Headers are title cased as opposed to all caps
// - NAME, VERSION, AUTHOR, COPYRIGHT, and GLOBAL OPTIONS sections are removed
// - Global options are displayed by name in the usage text
const CLI_APP_HELP_TEMPLATE = `Usage: {{if .UsageText }}{{.UsageText}}{{else}}{{.HelpName}} {{range $index, $option := .VisibleFlags}}[{{$option.GetName | PrefixedFirstFlagName}}] {{end}}{{if .Commands}}command [options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[args]{{end}}{{end}}{{if .Description}}

{{.Description}}{{end}}{{if .Commands}}

Commands:

{{range .Commands}}{{if not .HideHelp}}   {{join .Names ", "}}{{ "\t"}}{{.Usage}}{{ "\n" }}{{end}}{{end}}{{end}}`

// This version of the command help template has the following changes over the default ones:
// - Headers are title cased as opposed to all caps
// - NAME and CATEGORY sections are removed
const CLI_COMMAND_HELP_TEMPLATE = `Usage: {{if .UsageText}}{{.UsageText}}{{else}}{{.HelpName}}{{if .VisibleFlags}} [options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{else}}[args]{{end}}{{end}}{{if .Description}}

{{.Description}}{{end}}{{if .VisibleFlags}}

Options:

   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}`

// HelpPrinter that will wrap the text at
// HELP_TEXT_LINE_WIDTH characters, while preserving
// indentation and table tabs. Currently only works with
// tables delimited by `\t` and with only 2 columns.
func WrappedHelpPrinter(out io.Writer, templateString string, data interface{}) {
	funcMap := template.FuncMap{
		"join":                  strings.Join,
		"PrefixedFirstFlagName": PrefixedFirstFlagName,
	}

	templ := template.Must(template.New("help").Funcs(funcMap).Parse(templateString))
	rendered := bytes.NewBufferString("")
	err := templ.Execute(rendered, data)
	if err != nil {
		return
	}

	writer := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)
	for _, line := range strings.Split(rendered.String(), "\n") {
		indent := HelpTableAwareDetermineIndent(line, "\t+")
		wrappedLine := IndentAwareWrapText(line, HelpTextLineWidth, indent)
		fmt.Fprintln(writer, wrappedLine)
	}
	writer.Flush()
}

// regexp version of strings.SplitAfter.
// Similar functionality to regexp.Split, but returns the delimited strings
// with the trailing delimiter appended to it.
// Example:
//    re := regexp.MustCompile("\\s+")
//    text := "one two three"
//    out := RegexpSplitAfter(re, text)
//    out == ["one ", "two ", "three"]
func RegexpSplitAfter(re *regexp.Regexp, str string) []string {
	var out []string
	indexes := re.FindAllStringIndex(str, -1)
	if indexes == nil {
		return append(out, str)
	}
	cur := 0
	for _, index := range indexes {
		out = append(out, str[cur:index[1]])
		cur = index[1]
	}
	if cur != len(str) {
		out = append(out, str[cur:])
	}
	return out
}

// Wrap text to line width, while preserving any indentation
// Examples:
// in:
// text = `   exec\tExecute a command with temporary AWS credentials obtained by logging into Gruntwork Houston`
// lineWidth = 80
// indent = `       \t`
// out:
//    exec\tExecute a command with temporary AWS credentials obtained by
//        \tlogging into Gruntwork Houston
func IndentAwareWrapText(text string, lineWidth int, indent string) string {
	wrapped := ""
	re := regexp.MustCompile("\\s+")
	words := RegexpSplitAfter(re, text)
	if len(words) == 0 {
		return wrapped
	}
	// Keep on consuming words to current line until current line reaches
	// lineWidth, at which point we start a new line. Keep in mind that word is
	// word + whitespace to next word.
	wrapped = words[0]
	// NOTE: This tabLength is not exactly correct, as elastic
	//       tabstops work with cells which may cause the tabs
	//       to expand beyond 8 spaces. Nonetheless the goal
	//       is to avoid overflowing beyond the lineWidth and
	//       this should be a better approximation than 1
	//       char.
	tabLength := 8
	currentLineLength := TabAwareStringLength(wrapped, tabLength)
	for _, word := range words[1:] {
		wordLength := TabAwareStringLength(word, tabLength)
		trimmedWord := strings.TrimSpace(word)
		trimmedWordLength := TabAwareStringLength(trimmedWord, tabLength)
		if currentLineLength+trimmedWordLength > lineWidth {
			wrapped = strings.TrimRightFunc(wrapped, unicode.IsSpace)
			nextLine := indent + word
			wrapped += "\n" + nextLine
			currentLineLength = TabAwareStringLength(nextLine, tabLength)
		} else {
			wrapped += word
			currentLineLength += wordLength
		}
	}
	wrapped = strings.TrimRightFunc(wrapped, unicode.IsSpace)
	return wrapped
}

func TabAwareStringLength(text string, tabLength int) int {
	return len(strings.Replace(text, "\t", strings.Repeat(" ", tabLength), -1))
}

// Determine the indent string, accounting for textual tables.
// Assumes only two columns, and there is a clear delimiter for them, like in
// help text.
func HelpTableAwareDetermineIndent(text string, tableDelimiterRe string) string {
	// If we find a table, indent to second column
	tableRe := regexp.MustCompile(tableDelimiterRe)
	loc := tableRe.FindStringIndex(text)
	if loc != nil {
		return regexp.MustCompile("[^\\s]").ReplaceAllString(text[:loc[1]], " ")
	}

	// ... otherwise, indent one
	re := regexp.MustCompile("^\\s*")
	loc = re.FindStringIndex(text)
	if loc == nil {
		return ""
	}
	return text[loc[0]:loc[1]]
}

func PrefixedFirstFlagName(fullName string) string {
	names := strings.Split(fullName, ",")
	firstName := names[0]
	trimmedFirstName := strings.TrimSpace(firstName)
	if len(trimmedFirstName) == 1 {
		return "-" + trimmedFirstName
	} else {
		return "--" + trimmedFirstName
	}
}
