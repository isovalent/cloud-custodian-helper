package date_test

import (
	"c7n-helper/pkg/date"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseOrDefault(t *testing.T) {
	expected := time.Now()
	assert.Equal(t, expected, date.ParseOrDefault("", expected))
}
