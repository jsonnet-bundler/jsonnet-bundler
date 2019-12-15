package deps

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDependency(t *testing.T) {
	const testFolder = "test/jsonnet/foobar"
	err := os.MkdirAll(testFolder, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("test")

	tests := []struct {
		name string
		path string
		want *Dependency
	}{
		{
			name: "Empty",
			path: "",
			want: nil,
		},
		{
			name: "Invalid",
			path: "github.com/foo",
			want: nil,
		},
		{
			name: "local",
			path: testFolder,
			want: &Dependency{
				Source: Source{
					LocalSource: &Local{
						Directory: "test/jsonnet/foobar",
					},
				},
				Version: "",
			},
		},
	}
	for _, tt := range tests {
		_ = t.Run(tt.name, func(t *testing.T) {
			dependency := Parse("", tt.path)

			if tt.path == "" {
				assert.Nil(t, dependency)
			} else {
				assert.Equal(t, tt.want, dependency)
			}
		})
	}
}
