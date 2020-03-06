package shell

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFPromptUserForInputReturnsYesOnNonInteractive(t *testing.T) {
	t.Parallel()

	opts := NewShellOptions()
	opts.NonInteractive = true
	resp, err := FPromptUserForInput(os.Stdout, os.Stdin, "", opts)

	assert.Nil(t, err)
	assert.Equal(t, resp, "yes")
}

func TestFPromptUserForInputStripsInput(t *testing.T) {
	t.Parallel()

	opts := NewShellOptions()
	sout := ""
	fakeStdout := bytes.NewBufferString(sout)
	sin := "\t1.21 Gigawatts\t  \n"
	fakeStdin := bytes.NewBufferString(sin)
	resp, err := FPromptUserForInput(fakeStdout, fakeStdin, "", opts)

	assert.Nil(t, err)
	assert.Equal(t, resp, "1.21 Gigawatts")
}

func TestFPromptUserForInputAllowsEmptyString(t *testing.T) {
	t.Parallel()

	opts := NewShellOptions()
	sout := ""
	fakeStdout := bytes.NewBufferString(sout)
	sin := "\n"
	fakeStdin := bytes.NewBufferString(sin)
	resp, err := FPromptUserForInput(fakeStdout, fakeStdin, "", opts)

	assert.Nil(t, err)
	assert.Equal(t, resp, "")
}

func TestFPromptUserForInputPrintsOutPrompt(t *testing.T) {
	t.Parallel()

	opts := NewShellOptions()
	sout := ""
	fakeStdout := bytes.NewBufferString(sout)
	sin := "This is heavy\n"
	fakeStdin := bytes.NewBufferString(sin)
	_, err := FPromptUserForInput(fakeStdout, fakeStdin, "Great Scott!", opts)

	assert.Nil(t, err)
	assert.Contains(t, fakeStdout.String(), "Great Scott!")
}

func TestFPromptUserForYesNoPrintsOutPromptWithYN(t *testing.T) {
	t.Parallel()

	opts := NewShellOptions()
	sout := ""
	fakeStdout := bytes.NewBufferString(sout)
	sin := "y\n"
	fakeStdin := bytes.NewBufferString(sin)
	_, err := FPromptUserForYesNo(fakeStdout, fakeStdin, "Great Scott!", opts)

	assert.Nil(t, err)
	assert.Contains(t, fakeStdout.String(), "Great Scott! (y/n)")
}

var yesNoPromptTests = []struct {
	in  string
	out bool
}{
	{"y", true},
	{"YEs", true},
	{"Y", true},
	{"YES", true},
	{"yes     ", true},
	{"    yes", true},
	{"\tyes", true},
	{"yes\t", true},
	{"ye", false},
	{"", false},
	{"delorean", false},
	{"yes no", false},
}

func TestFPromptUserForYesNo(t *testing.T) {
	for _, tt := range yesNoPromptTests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()

			opts := NewShellOptions()
			sout := ""
			fakeStdout := bytes.NewBufferString(sout)
			sin := tt.in + "\n"
			fakeStdin := bytes.NewBufferString(sin)
			resp, err := FPromptUserForYesNo(fakeStdout, fakeStdin, "Great Scott!", opts)

			assert.Nil(t, err)
			assert.Equal(t, resp, tt.out)
		})
	}
}
