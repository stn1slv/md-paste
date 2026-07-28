package service

import (
	"errors"
	"testing"

	"github.com/stn1slv/md-paste/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const htmlTable = "<table><tr><th>H1</th></tr><tr><td>D1</td></tr></table>"

func readHTML() (models.ClipboardContent, error) {
	return models.ClipboardContent{
		RawHTML:     htmlTable,
		ContentType: models.ContentTypeHTML,
	}, nil
}

const expectedTableMarkdown = "| H1 |\n| --- |\n| D1 |"

func TestConvertEmptyClipboard(t *testing.T) {
	read := func() (models.ClipboardContent, error) {
		return models.ClipboardContent{ContentType: models.ContentTypeNone}, nil
	}
	writeCalled := false
	write := func(string) error {
		writeCalled = true
		return nil
	}

	outcome, _, err := Convert(read, write, Hooks{})
	require.NoError(t, err)
	assert.Equal(t, OutcomeEmpty, outcome)
	assert.False(t, writeCalled, "write must not be called on an empty clipboard")
}

func TestConvertWritesMarkdown(t *testing.T) {
	var captured string
	write := func(md string) error {
		captured = md
		return nil
	}

	outcome, doc, err := Convert(readHTML, write, Hooks{})
	require.NoError(t, err)
	assert.Equal(t, OutcomeConverted, outcome)
	assert.Equal(t, expectedTableMarkdown, captured)
	assert.Equal(t, expectedTableMarkdown, doc.Content)
}

func TestConvertOnRawContentHook(t *testing.T) {
	var rawSeen models.ClipboardContent
	hooks := Hooks{
		OnRawContent: func(c models.ClipboardContent) error {
			rawSeen = c
			return nil
		},
	}

	outcome, _, err := Convert(readHTML, func(string) error { return nil }, hooks)
	require.NoError(t, err)
	assert.Equal(t, OutcomeConverted, outcome)
	assert.Equal(t, htmlTable, rawSeen.RawHTML)
}

func TestConvertOnRawContentError(t *testing.T) {
	sentinel := errors.New("disk full")
	hooks := Hooks{
		OnRawContent: func(models.ClipboardContent) error { return sentinel },
	}

	_, _, err := Convert(readHTML, func(string) error { return nil }, hooks)
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}

func TestConvertSinkOverridesWrite(t *testing.T) {
	writeCalled := false
	write := func(string) error {
		writeCalled = true
		return nil
	}
	var sunk string
	hooks := Hooks{Sink: func(md string) error {
		sunk = md
		return nil
	}}

	outcome, _, err := Convert(readHTML, write, hooks)
	require.NoError(t, err)
	assert.Equal(t, OutcomeConverted, outcome)
	assert.False(t, writeCalled, "write must not be called when Sink is set")
	assert.Equal(t, expectedTableMarkdown, sunk)
}

func TestConvertReadError(t *testing.T) {
	sentinel := errors.New("clipboard unavailable")
	read := func() (models.ClipboardContent, error) {
		return models.ClipboardContent{}, sentinel
	}

	_, _, err := Convert(read, func(string) error { return nil }, Hooks{})
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}

func TestConvertWriteError(t *testing.T) {
	sentinel := errors.New("clipboard write failed")
	write := func(string) error { return sentinel }

	_, _, err := Convert(readHTML, write, Hooks{})
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}
