package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashDataStore(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		s := newDataStore()
		_, ok := s.get("asd")
		assert.False(t, ok)

		s.delete("foo")
	})

	t.Run("set get", func(t *testing.T) {
		s := newDataStore()
		s.store("foo", "123")

		v, ok := s.get("foo")
		assert.True(t, ok)
		assert.Equal(t, "123", v)
	})

	t.Run("set delete get", func(t *testing.T) {
		s := newDataStore()
		s.store("foo", "123")

		s.delete("foo")

		_, ok := s.get("foo")
		assert.False(t, ok)
	})
}