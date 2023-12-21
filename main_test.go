package main

import (
	"testing"

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
