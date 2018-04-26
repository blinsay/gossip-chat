package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChat_since(t *testing.T) {
	c := &chat{}
	c.send("dr nick", "hi everybody")
	c.send("everybody", "hi dr nick")
	c.send("dr nick", "hi everybody")
	seen := c.send("everybody", "hi dr nick")
	c.send("dr nick", "hi everybody")
	c.send("everybody", "hi dr nick")

	assert.Equal(t, []string{"hi everybody", "hi dr nick"}, messageText(c.since(seen)))
}

func TestChat_merge(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		c1 := &chat{}
		c2 := &chat{}
		c2.send("dr nick", "hi everybody")

		c1.merge(c2)

		assert.Equal(t, c1, c2)
	})

	t.Run("idempotent", func(t *testing.T) {
		c := &chat{}
		c.send("dr nick", "hi everybody")
		c.send("everybody", "hi dr nick")

		c.merge(c)

		assert.Equal(t, []string{"hi everybody", "hi dr nick"}, messageText(c))
	})

	t.Run("merges in order", func(t *testing.T) {
		c1 := &chat{}
		c1.send("dr nick 0", "hi everybody")
		c1.send("everybody", "hi dr nick")
		c2 := &chat{}
		c2.send("dr nick 1", "hi everybody else")

		c1.merge(c2)

		assert.Equal(t, []string{"hi everybody", "hi everybody else", "hi dr nick"}, messageText(c1))
	})
}

func messageText(c *chat) []string {
	msgs := make([]string, len(c.Messages))
	for i := range c.Messages {
		msgs[i] = c.Messages[i].Txt
	}
	return msgs
}
