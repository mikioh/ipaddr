// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

// +build go1.2

package ipaddr

import (
	"encoding"
	"net"
)

var (
	_ encoding.TextMarshaler   = &IPv4{}
	_ encoding.TextMarshaler   = &IPv6{}
	_ encoding.TextUnmarshaler = &IPv4{}
	_ encoding.TextUnmarshaler = &IPv6{}
)

// MarshalText implements the MarshalText method of
// encoding.TextMarshaler interface.
func (p *IPv4) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalText implements the UnmarshalText method of
// encoding.TextUnmarshaler interface.
func (p *IPv4) UnmarshalText(text []byte) error {
	s := string(text)
	_, ipn, err := net.ParseCIDR(s)
	if err != nil {
		return err
	}
	nbits, _ := ipn.Mask.Size()
	return p.Set(ipn.IP, nbits)
}

// MarshalText implements the MarshalText method of
// encoding.TextMarshaler interface.
func (p *IPv6) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalText implements the UnmarshalText method of
// encoding.TextUnmarshaler interface.
func (p *IPv6) UnmarshalText(text []byte) error {
	s := string(text)
	_, ipn, err := net.ParseCIDR(s)
	if err != nil {
		return err
	}
	nbits, _ := ipn.Mask.Size()
	return p.Set(ipn.IP, nbits)
}
