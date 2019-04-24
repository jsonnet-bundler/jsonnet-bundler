// +build integration

package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCommand(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "jb-init")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempDir)

	code := initCommand(tempDir)
	assert.Equal(t, 0, code)
}
