// Copyright 2015 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr_test

import (
	"errors"
	"net"
	"reflect"
	"testing"

	"github.com/mikioh/ipaddr"
)

func toPosition(s1, s2 string) *ipaddr.Position {
	if s1 == "" || s2 == "" {
		return nil
	}
	return &ipaddr.Position{IP: net.ParseIP(s1), Prefix: *toPrefix(s2)}
}

var cursorFirstLastIPTests = []struct {
	in          []string
	first, last net.IP
}{
	// IPv4 prefixes
	{
		[]string{"0.0.0.0/0", "255.255.255.255/32"},
		net.ParseIP("0.0.0.0"), net.ParseIP("255.255.255.255"),
	},
	{
		[]string{"192.168.0.0/32", "192.168.0.1/32", "192.168.0.2/32", "192.168.0.3/32", "192.168.4.0/24", "192.168.0.0/32", "192.168.0.1/32"},
		net.ParseIP("192.168.0.0"), net.ParseIP("192.168.4.255"),
	},
	{
		[]string{"192.168.0.1/32"},
		net.ParseIP("192.168.0.1"), net.ParseIP("192.168.0.1"),
	},

	// IPv6 prefixes
	{
		[]string{"::/0", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"},
		net.ParseIP("::"), net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"),
	},
	{
		[]string{"2001:db8::/64", "2001:db8:0:1::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64", "2001:db8:0:4::/64", "2001:db8::/64", "2001:db8::1/64"},
		net.ParseIP("2001:db8::"), net.ParseIP("2001:db8:0:4:ffff:ffff:ffff:ffff"),
	},
	{
		[]string{"2001:db8::1/128"},
		net.ParseIP("2001:db8::1"), net.ParseIP("2001:db8::1"),
	},

	// Mixed prefixes
	{
		[]string{"192.168.0.1/32", "2001:db8::1/64", "192.168.255.0/24"},
		net.ParseIP("192.168.0.1"), net.ParseIP("2001:db8::ffff:ffff:ffff:ffff"),
	},
}

func TestCursorFirstLastIP(t *testing.T) {
	for i, tt := range cursorFirstLastIPTests {
		in := toPrefixes(tt.in)
		c := ipaddr.NewCursor(in)
		fpos, lpos := c.First(), c.Last()
		if !tt.first.Equal(fpos.IP) || !tt.last.Equal(lpos.IP) {
			t.Errorf("#%d: got %v, %v; want %v, %v", i, fpos.IP, lpos.IP, tt.first, tt.last)
		}
	}
}

var cursorFirstLastPrefixTests = []struct {
	in          []string
	first, last *ipaddr.Prefix
}{
	// IPv4 prefixes
	{
		[]string{"0.0.0.0/0", "255.255.255.255/32"},
		toPrefix("0.0.0.0/0"), toPrefix("255.255.255.255/32"),
	},
	{
		[]string{"192.168.0.0/32", "192.168.0.1/32", "192.168.0.2/32", "192.168.0.3/32", "192.168.4.0/24", "192.168.0.0/32", "192.168.0.1/32"},
		toPrefix("192.168.0.0/32"), toPrefix("192.168.4.0/24"),
	},

	// IPv6 prefixes
	{
		[]string{"::/0", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"},
		toPrefix("::/0"), toPrefix("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"),
	},
	{
		[]string{"2001:db8::/64", "2001:db8:0:1::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64", "2001:db8:0:4::/64", "2001:db8::/64", "2001:db8::1/64"},
		toPrefix("2001:db8::/64"), toPrefix("2001:db8:0:4::/64"),
	},

	// Mixed prefixes
	{
		[]string{"192.168.0.1/32", "2001:db8::1/64", "192.168.255.0/24"},
		toPrefix("192.168.0.1/32"), toPrefix("2001:db8::/64"),
	},
}

func TestCursorFirstLastPrefix(t *testing.T) {
	for i, tt := range cursorFirstLastPrefixTests {
		in := toPrefixes(tt.in)
		c := ipaddr.NewCursor(in)
		fpos, lpos := c.First(), c.Last()
		if !tt.first.Equal(&fpos.Prefix) || !tt.last.Equal(&lpos.Prefix) {
			t.Errorf("#%d: got %v, %v; want %v, %v", i, fpos.Prefix, lpos.Prefix, tt.first, tt.last)
		}
	}
}

var cursorNextTests = []struct {
	ps       []string
	in, want *ipaddr.Position
}{
	// IPv4 prefixes
	{
		[]string{"192.168.0.0/24"},
		toPosition("192.168.0.255", "192.168.0.0/24"), nil,
	},
	{
		[]string{"192.168.0.0/24", "192.168.1.0/24"},
		toPosition("192.168.0.255", "192.168.0.0/24"), toPosition("192.168.1.0", "192.168.1.0/24"),
	},

	// IPv6 prefixes
	{
		[]string{"2001:db8::/64"},
		toPosition("2001:db8::ffff:ffff:ffff:ffff", "2001:db8::/64"), nil,
	},
	{
		[]string{"2001:db8::/64", "2001:db8:1::/64"},
		toPosition("2001:db8::ffff:ffff:ffff:ffff", "2001:db8::/64"), toPosition("2001:db8:1::", "2001:db8:1::/64"),
	},

	// Mixed prefixes
	{
		[]string{"192.168.0.0/24", "2001:db8::/64"},
		toPosition("2001:db8::ffff:ffff:ffff:ffff", "2001:db8::/64"), nil,
	},
	{
		[]string{"192.168.0.0/24", "::/64"},
		toPosition("192.168.0.255", "192.168.0.0/24"), nil,
	},
	{
		[]string{"192.168.0.0/24", "2001:db8::/64"},
		toPosition("192.168.0.255", "192.168.0.0/24"), toPosition("2001:db8::", "2001:db8::/64"),
	},
	{
		[]string{"192.168.0.0/24", "::/64"},
		toPosition("::ffff:ffff:ffff:ffff", "::/64"), toPosition("192.168.0.0", "192.168.0.0/24"),
	},
}

func TestCursorNext(t *testing.T) {
	for i, tt := range cursorNextTests {
		ps := toPrefixes(tt.ps)
		c := ipaddr.NewCursor(ps)
		if err := c.Set(tt.in); err != nil {
			t.Fatal(err)
		}
		out := c.Next()
		if !reflect.DeepEqual(out, tt.want) {
			t.Errorf("#%d: got %v; want %v", i, out, tt.want)
		}
	}
}

var cursorPosTests = []struct {
	ps []string
	in *ipaddr.Position
	error
}{
	// IPv4 prefixes
	{
		[]string{"192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/24"},
		toPosition("192.168.1.1", "192.168.1.0/24"),
		nil,
	},
	{
		[]string{"192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/24"},
		toPosition("192.168.3.1", "192.168.1.0/24"),
		errors.New("should fail"),
	},
	{
		[]string{"192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/24"},
		toPosition("192.168.1.1", "192.168.3.0/24"),
		errors.New("should fail"),
	},

	// IPv6 prefixes
	{
		[]string{"2001:db8::/64", "2001:db8:1::/64", "2001:db8:2::/64"},
		toPosition("2001:db8:1::1", "2001:db8:1::/64"),
		nil,
	},
	{
		[]string{"2001:db8::/64", "2001:db8:1::/64", "2001:db8:2::/64"},
		toPosition("2001:db8:3::1", "2001:db8:1::/64"),
		errors.New("should fail"),
	},
	{
		[]string{"2001:db8::/64", "2001:db8:1::/64", "2001:db8:2::/64"},
		toPosition("2001:db8:1::1", "2001:db8:3::/64"),
		errors.New("should fail"),
	},

	// Mixed prefixes
	{
		[]string{"192.168.0.0/24", "2001:db8::/64"},
		toPosition("192.168.0.1", "192.168.0.0/24"),
		nil,
	},
	{
		[]string{"192.168.0.0/24", "2001:db8::/64"},
		toPosition("2001:db8::1", "2001:db8::/64"),
		nil,
	},
	{
		[]string{"192.168.0.0/24", "2001:db8::/64"},
		toPosition("2001:db8::1", "192.168.0.0/24"),
		errors.New("should fail"),
	},
	{
		[]string{"192.168.0.0/24", "2001:db8::/64"},
		toPosition("192.168.0.1", "2001:db8::/64"),
		errors.New("should fail"),
	},
}

func TestCursorPos(t *testing.T) {
	for i, tt := range cursorPosTests {
		ps := toPrefixes(tt.ps)
		c := ipaddr.NewCursor(ps)
		err := c.Set(tt.in)
		if err != nil && tt.error == nil {
			t.Errorf("#%d: got %v; want %v", i, err, tt.error)
		}
		if err != nil {
			continue
		}
		if !reflect.DeepEqual(c.Pos(), tt.in) {
			t.Errorf("#%d: got %v; want %v", i, c.Pos(), tt.in)
		}
	}
}

var newCursorTests = []struct {
	in []string
}{
	// IPv4 prefixes
	{
		[]string{"192.168.0.0/32", "192.168.0.1/32", "192.168.0.2/32", "192.168.0.3/32", "192.168.4.0/24", "192.168.0.0/32", "192.168.0.1/32"},
	},

	// IPv6 prefixes
	{
		[]string{"2001:db8::/64", "2001:db8:0:1::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64", "2001:db8:0:4::/64", "2001:db8::/64", "2001:db8::1/64"},
	},

	// Mixed prefixes
	{
		[]string{"192.168.0.0/32", "192.168.0.1/32", "192.168.0.2/32", "192.168.0.3/32", "192.168.4.0/24", "192.168.0.0/32", "192.168.0.1/32", "2001:db8::/64", "2001:db8:0:1::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64", "2001:db8:0:4::/64", "2001:db8::/64", "2001:db8::1/64"},
	},
}

func TestNewCursor(t *testing.T) {
	for i, tt := range newCursorTests {
		in, orig := toPrefixes(tt.in), toPrefixes(tt.in)
		ipaddr.NewCursor(in)
		if !reflect.DeepEqual(in, orig) {
			t.Errorf("#%d: %v is corrupted; want %v", i, in, orig)
		}
	}

	ipaddr.NewCursor(nil)
}
