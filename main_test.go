package main

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestListGophers(t *testing.T) {
	_, err := ListGophers()
	assert.NoError(t, err)
}

func TestGetGopher(t *testing.T) {
	_, err := GetGopher("gandalf", false)
	assert.NoError(t, err)
	_, err = GetGopher("", true)
	assert.NoError(t, err)
}

func TestParseCommand(t *testing.T) {
	tests := []struct {
		testName              string
		command               string
		expectThrowPattern    *regexp.Regexp
		expectContentPattern  *regexp.Regexp
		expectFileNamePattern *regexp.Regexp
	}{
		{
			testName:           "invalid command",
			command:            "!asdf",
			expectThrowPattern: regexp.MustCompile(`Unknown command: "!asdf"`),
		},
		{
			testName:              "any gopher",
			command:               "!gopher",
			expectContentPattern:  regexp.MustCompile(`Random gopher go!`),
			expectFileNamePattern: regexp.MustCompile(`.*`),
		},
		{
			testName:              "gandalf",
			command:               "!gopher gandalf",
			expectContentPattern:  regexp.MustCompile(`Introducing Ser Gandalf`),
			expectFileNamePattern: regexp.MustCompile(`.*`),
		},
		{
			testName:              "fire gopher",
			command:               "!gopher fire-gopher",
			expectContentPattern:  regexp.MustCompile(`Introducing Ser Fire-Gopher`),
			expectFileNamePattern: regexp.MustCompile(`.*`),
		},
		{
			testName:             "not a gopher",
			command:              "!gopher not-a-gopher",
			expectContentPattern: regexp.MustCompile(`I don't know this "not-a-gopher" guy`),
		},
		{
			testName:             "get gopher list",
			command:              "!gophers",
			expectContentPattern: regexp.MustCompile(`([-a-z]+\n)+`),
		},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			input := &discordgo.MessageCreate{
				Message: &discordgo.Message{Content: tc.command},
			}
			output, files, err := ParseCommand(input)
			if tc.expectThrowPattern != nil {
				assert.Error(t, err)
				assert.True(t, tc.expectThrowPattern.MatchString(fmt.Sprint(err)), fmt.Sprintf("Unexpected error message, expected:\n%v\nGot:\n%v", tc.expectThrowPattern, err))
			} else {
				assert.NoError(t, err)
				assert.True(t, tc.expectContentPattern.MatchString(output.Content))
				if tc.expectFileNamePattern != nil {
					assert.True(t, tc.expectFileNamePattern.MatchString(output.File.Name))
					assert.NotNil(t, files)
				}
			}
			for _, file := range files {
				file.Close()
			}
		})
	}
}
