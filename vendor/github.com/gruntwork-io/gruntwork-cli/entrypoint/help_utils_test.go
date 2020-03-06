package entrypoint

import (
	"regexp"

	"github.com/stretchr/testify/assert"
	"testing"
)

var splitAfterTests = []struct {
	in  string
	out []string
}{
	{"one two three", []string{"one ", "two ", "three"}},
	{"one\ttwo\tthree", []string{"one\t", "two\t", "three"}},
	{"one\n two    three\t\n ", []string{"one\n ", "two    ", "three\t\n "}},
	{"onetwothree\n", []string{"onetwothree\n"}},
	{"\nonetwothree", []string{"\n", "onetwothree"}},
}

func TestRegexpSplitAfterDelimiter(t *testing.T) {
	for _, tt := range splitAfterTests {
		t.Run(tt.in, func(t *testing.T) {
			re := regexp.MustCompile("\\s+")
			assert.Equal(t, RegexpSplitAfter(re, tt.in), tt.out)
		})
	}
}

var determineIndentTests = []struct {
	in    string
	delim string
	out   string
}{
	{"   o three", "\t", "   "},
	{"\to  three", "\t", "\t"},
	{"o three", "\t", ""},
	{"  one\ttwo", "\t", "     \t"},
	{"  \ttwo", "\t", "  \t"},
	{"  hello|world", "\\|", "        "},
	{
		"   exec\tExecute a command with temporary AWS credentials obtained by logging into Gruntwork Houston",
		"\t",
		"       \t",
	},
}

func TestHelpTableAwareDetermineIndent(t *testing.T) {
	for _, tt := range determineIndentTests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, HelpTableAwareDetermineIndent(tt.in, tt.delim), tt.out)
		})
	}
}

var wrapTextTests = []struct {
	in        string
	out       string
	indent    string
	lineWidth int
}{
	{
		"    Great Scott!",
		"    Great\n    Scott!",
		"    ",
		15,
	},
	{
		"You made a time machine out of a Delorean!?",
		"You made a time\nmachine out of\na Delorean!?",
		"",
		15,
	},
	{
		"  fc\tThe box that",
		"  fc\tThe\n    \tbox\n    \tthat",
		"    \t",
		15,
	},
	{
		"   exec\tExecute a command with temporary AWS credentials obtained by logging into Gruntwork Houston",
		"   exec\tExecute a command with temporary AWS credentials obtained by\n       \tlogging into Gruntwork Houston",
		"       \t",
		80,
	},
}

func TestIndentAwareWrapText(t *testing.T) {
	for _, tt := range wrapTextTests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(
				t,
				IndentAwareWrapText(tt.in, tt.lineWidth, tt.indent),
				tt.out)
		})
	}
}
