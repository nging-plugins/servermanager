package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgsSlice(t *testing.T) {
	r := ParseArgsSlice("-c=config.yml\n-b='Y'")
	assert.Equal(t, []string{`-c`, `config.yml`, `-b`, `Y`}, r)

	r = ParseArgsSlice("-c config.yml\n-b 'Y'")
	assert.Equal(t, []string{`-c`, `config.yml`, `-b`, `Y`}, r)
}
