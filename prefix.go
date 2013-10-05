// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

// Package ipaddr provides basic functions for the manipulation of IP
// address prefixes and subsequent addresses as described in RFC 4632
// and RFC 4291.
package ipaddr

import (
	"errors"
	"math/big"
	"net"
)

var errInvalidArgument = errors.New("invalid arugument")

// A Prefix represents an IP address prefix.
type Prefix interface {
	// Contains reports whether the prefix includes the given ip.
	Contains(ip net.IP) bool

	// Overlaps reports whether the prefix overlaps with the given
	// prefix.
	Overlaps(prefix Prefix) bool

	// Equal reports whether the prefix and the given prefix are
	// equal.
	Equal(prefix Prefix) bool

	// String returns the string form of the prefix.
	String() string

	// Len returns the length of the prefix in bits.
	Len() int

	// NumAddr returns the number of addresses possible in the
	// prefix.
	NumAddr() *big.Int

	// Bits returns the nbits bit sequence extracted from the
	// prefix at position pos.
	Bits(pos, nbits int) uint32

	// Addr returns the address of the prefix.
	Addr() net.IP

	// LastAddr returns the last address in the address range of
	// the prefix. It returns the address of the prefix when the
	// prefix contains only one address.
	LastAddr() net.IP

	// Hostmask returns the host mask, the inverse mask of the
	// prefix's network mask.
	Hostmask() net.IPMask

	// Netmask returns the network mask of the prefix.
	Netmask() net.IPMask

	// Hosts returns the list of addresses that are assignable to
	// hosts, nodes that are not routers or other intermediate
	// systems, beginning with the given address begin for the
	// prefix neither 0.0.0.0/0 nor ::/0.
	//
	// Note that it will take a bit long time when the prefix is a
	// shorter one, and it also does not distinguish any multicast
	// addresses correctly for now.
	Hosts(begin net.IP) []net.IP

	// HostIter generates and returns the iterator that iterates
	// over the list of addresses that are assignable to hosts,
	// beginning with the given address's next address.
	//
	// Note that it does not identify host-assignable addresses
	// when the prefix is 0.0.0.0/0 or ::/0, and it also does not
	// distinguish any multicast addresses correctly for now.
	HostIter(first net.IP) <-chan net.IP

	// Subnets returns the list of prefixes that are splitted from
	// the prefix, into small address blocks by nbits which
	// represents a number of subnetworks in power of 2 notation.
	Subnets(nbits int) []Prefix

	// SubnetIter generates and returns the iterator that iterates
	// over the list of prefixes that are splitted from the
	// prefix, into small address blocks by nbits which represents
	// a number of subnetworks in power of 2 notation.
	SubnetIter(nbits int) <-chan Prefix

	// Exclude returns the list of prefixes that do not contain
	// the given prefix.
	Exclude(prefix Prefix) []Prefix

	// Set replaces the existing address and prefix length of the
	// prefix with ip and nbits.
	Set(ip net.IP, nbits int) error

	// MarshalText returns the UTF-8-encoded text form of the
	// prefix.
	MarshalText() ([]byte, error)

	// UnmarshalText replaces the existing address and prefix
	// length of the prefix with text.
	UnmarshalText(text []byte) error
}

// NewPrefix returns a new Prefix.
func NewPrefix(ip net.IP, nbits int) (Prefix, error) {
	if ipv4 := ip.To4(); ipv4 != nil && 0 <= nbits && nbits <= IPv4PrefixLen {
		return newIPv4(ipToIPv4Int(ipv4), byte(nbits)), nil
	} else if ipv6 := ip.To16(); ipv6 != nil && ipv6.To4() == nil && 0 <= nbits && nbits <= IPv6PrefixLen {
		return newIPv6(ipToIPv6Int(ipv6), byte(nbits)), nil
	}
	return nil, errInvalidArgument
}

// ComparePrefix returns an integer comparing two prefixes. The result
// will be 0 if a == b, -1 if a < b, and +1 if a > b.
func ComparePrefix(a, b Prefix) int {
	switch a.(type) {
	case *IPv4:
		return ipv4ComparePrefix(a.(*IPv4), b.(*IPv4))
	case *IPv6:
		return ipv6ComparePrefix(a.(*IPv6), b.(*IPv6))
	default:
		panic("unknown address family")
	}
}

// SummaryPrefix tries to find out a shortest common prefix for the
// given prefixes. It returns nil when no suitable prefix is found.
func SummaryPrefix(prefixes []Prefix) Prefix {
	if len(prefixes) == 0 {
		return nil
	} else if len(prefixes) == 1 {
		return prefixes[0]
	}
	switch prefixes[0].(type) {
	case *IPv4:
		p := ipv4SummaryPrefix(prefixes)
		if p == nil {
			return nil
		}
		return p
	case *IPv6:
		p := ipv6SummaryPrefix(prefixes)
		if p == nil {
			return nil
		}
		return p
	default:
		return nil
	}
}
