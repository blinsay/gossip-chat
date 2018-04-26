package main

import (
	"sort"
	"sync"
)

// a lamport clock.
//
// see: https://en.wikipedia.org/wiki/Lamport_timestamps
type clock uint64

func (c clock) increment() clock {
	return c + 1
}

func (c clock) update(o clock) clock {
	if c > o {
		return c
	}
	return o
}

// a chat is a partially ordered sequence of messages, implemented as a CRDT.
// not safe for access from multiple goroutines.
//
// messages are ordered by clock, with ties broken by sender ID (whomst). a chat
// may not contain multiple messages with the same clock and sender ID - when
// conflicts arise, the first message with the given (clock, sender ID) pair
// should be kept.
type chat struct {
	mu sync.RWMutex

	Messages []message `json:"messages"`
}

// a single message in a chat
type message struct {
	Clock  clock  `json:"clock"`
	Whomst string `json:"whomst"`
	Txt    string `json:"txt"`
}

func (c *chat) lastMessageAt() clock {
	if len(c.Messages) == 0 {
		return 0
	}
	return c.Messages[len(c.Messages)-1].Clock
}

// return all new messages since a given timestamp
func (c *chat) since(t clock) *chat {
	c.mu.RLock()
	defer c.mu.RUnlock()

	i := sort.Search(len(c.Messages), func(i int) bool {
		return c.Messages[i].Clock > t
	})

	since := &chat{}
	since.Messages = append(since.Messages, c.Messages[i:]...)
	return since
}

// send a new message to this chat
func (c *chat) send(who, msg string) clock {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.Messages) == 0 {
		c.Messages = append(c.Messages, message{Whomst: who, Txt: msg, Clock: 1})
		return 0
	}

	nextClock := c.Messages[len(c.Messages)-1].Clock.increment()
	c.Messages = append(c.Messages, message{Whomst: who, Txt: msg, Clock: nextClock})
	return nextClock
}

// merge another chat history into this one.
func (c *chat) merge(other *chat) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// FIXME(benl): we're doing a sort and then removing consecutive duplicates.
	// there's definitely a smarter way to do this since slices should already be
	// (mostly) sorted, but whatever.

	// append
	c.Messages = append(c.Messages, other.Messages...)
	// sort by clock and by person
	sort.SliceStable(c.Messages, func(i, j int) bool {
		m1, m2 := c.Messages[i], c.Messages[j]
		if m1.Clock == m2.Clock {
			return m1.Whomst < m2.Whomst
		}
		return m1.Clock < m2.Clock
	})
	// remove duplicates
	for i := len(c.Messages) - 1; i > 0; i-- {
		current, prev := c.Messages[i], c.Messages[i-1]
		if current.Clock == prev.Clock && current.Whomst == prev.Whomst {
			c.Messages = append(c.Messages[:i], c.Messages[i+1:]...)
		}
	}
}
