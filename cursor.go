// Copyright 2015 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr

import (
	"errors"
	"net"
)

// A Position represents a position on a cursor.
type Position struct {
	IP     net.IP // IP address
	Prefix Prefix // IP address prefix
}

// A Cursor repesents a movable indicator on multiple prefixes.
type Cursor struct {
	curr, last ipv6Int

	pi int
	ps []Prefix // IP address prefixes
}

func (c *Cursor) next() {
	c.pi++
	ip := c.ps[c.pi].IP.To16()
	c.curr = ipToIPv6Int(ip)
	if ip.To4() != nil {
		c.last = c.ps[c.pi].lastIPv4MappedInt()
	}
	if ip.To16() != nil && ip.To4() == nil {
		c.last = c.ps[c.pi].lastIPv6Int()
	}
}

func (c *Cursor) set(pi int, ip net.IP) {
	c.pi = pi
	c.curr = ipToIPv6Int(ip.To16())
	if ip.To4() != nil {
		c.last = c.ps[c.pi].lastIPv4MappedInt()
	}
	if ip.To16() != nil && ip.To4() == nil {
		c.last = c.ps[c.pi].lastIPv6Int()
	}
}

// First returns the first position on the cursor c.
func (c *Cursor) First() *Position {
	return &Position{IP: c.ps[0].IP, Prefix: c.ps[0]}
}

// Last returns the end position on the cursor c.
func (c *Cursor) Last() *Position {
	return &Position{IP: c.ps[len(c.ps)-1].Last(), Prefix: c.ps[len(c.ps)-1]}
}

// List returns a list of prefixes on the cursor c.
func (c *Cursor) List() []Prefix {
	return c.ps
}

// Next returns the next position on the cursor c.
// It returns nil at the end on the cursor c.
func (c *Cursor) Next() *Position {
	n := c.curr.cmp(&c.last)
	if n == 0 {
		if c.pi == len(c.ps)-1 {
			return nil
		}
		c.next()
	} else {
		c.curr.incr()
	}
	return c.Pos()
}

// Pos returns the current postion on the cursor c.
func (c *Cursor) Pos() *Position {
	return &Position{IP: c.curr.ip(), Prefix: c.ps[c.pi]}
}

// Set sets the current postion on the cursor c to pos.
func (c *Cursor) Set(pos *Position) error {
	if pos == nil {
		return errors.New("invalid position")
	}
	pi := -1
	for i, p := range c.ps {
		if p.Equal(&pos.Prefix) {
			pi = i
			break
		}
	}
	if pi == -1 || !c.ps[pi].Contains(pos.IP) {
		return errors.New("position out of range")
	}
	c.set(pi, pos.IP.To16())
	return nil
}

// NewCursor returns a new cursor.
func NewCursor(ps []Prefix) *Cursor {
	if len(ps) == 0 {
		return nil
	}
	ps = sortAndDedup(ps, false)
	if len(ps) == 0 {
		return nil
	}
	c := Cursor{ps: ps}
	c.set(0, c.ps[0].IP.To16())
	return &c
}
