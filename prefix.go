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
	"sort"
	"strconv"
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

	// MarshalBinary returns the BGP NLRI binary form of the
	// prefix.
	MarshalBinary() ([]byte, error)

	// UnmarshalBinary replaces the existing address and prefix
	// length of the prefix with data.
	UnmarshalBinary(data []byte) error

	// MarshalText returns the UTF-8-encoded text form of the
	// prefix.
	MarshalText() ([]byte, error)

	// UnmarshalText replaces the existing address and prefix
	// length of the prefix with text.
	UnmarshalText(text []byte) error
}

// MustParsePrefix return a parsed Prefix. Panic if error occur.
func MustParsePrefix(s string) Prefix {
	prefix, err := ParsePrefix(s)
	if err != nil {
		panic(`ipaddr: ParsePrefix(` + strconv.Quote(s) + `): ` + err.Error())
	}
	return prefix
}

// ParsePrefix return a parsed Prefix.
func ParsePrefix(s string) (Prefix, error) {
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, errInvalidArgument
	}

	nbits, _ := ipnet.Mask.Size()
	return NewPrefix(ipnet.IP, nbits)
}

// NewPrefix returns a new Prefix.
func NewPrefix(ip net.IP, nbits int) (Prefix, error) {
	if ipv4 := ip.To4(); ipv4 != nil && 0 <= nbits && nbits <= IPv4PrefixLen {
		return newIPv4(ipToIPv4Int(ipv4), nbits), nil
	} else if ipv6 := ip.To16(); ipv6 != nil && ipv6.To4() == nil && 0 <= nbits && nbits <= IPv6PrefixLen {
		return newIPv6(ipToIPv6Int(ipv6), nbits), nil
	}
	return nil, errInvalidArgument
}

// Compare returns an integer comparing two prefixes. The result will
// be 0 if a == b, -1 if a < b, and +1 if a > b.
func Compare(a, b Prefix) int {
	switch a := a.(type) {
	case *IPv4:
		b, ok := b.(*IPv4)
		if !ok {
			return -1
		}
		return a.compare(b)
	case *IPv6:
		b, ok := b.(*IPv6)
		if !ok {
			return -1
		}
		return a.compare(b)
	default:
		return -1
	}
}

// Supernet finds out a shortest common address prefix for the given
// prefixes. It returns nil when no suitable prefix is found.
func Supernet(prefixes []Prefix) Prefix {
	if len(prefixes) == 0 {
		return nil
	}
	switch prefixes[0].(type) {
	case *IPv4:
		subs := byAddrFamily(prefixes).ipv4Only()
		if l := len(subs); l == 0 {
			return nil
		} else if l == 1 {
			return subs[0]
		}
		return supernetIPv4(subs)
	case *IPv6:
		subs := byAddrFamily(prefixes).ipv6Only()
		if l := len(subs); l == 0 {
			return nil
		} else if l == 1 {
			return subs[0]
		}
		return supernetIPv6(subs)
	default:
		return nil
	}
}

// Aggregate aggregates the given prefixes and returns a list of
// aggregated address prefixes.
func Aggregate(prefixes []Prefix) []Prefix {
	if len(prefixes) == 0 {
		return nil
	}
	switch prefixes[0].(type) {
	case *IPv4:
		subs := sortAndDedup(prefixes)
		if l := len(subs); l == 0 {
			return nil
		} else if l == 1 {
			return subs
		}
		return aggregateIPv4(subs)
	case *IPv6:
		subs := sortAndDedup(prefixes)
		if l := len(subs); l == 0 {
			return nil
		} else if l == 1 {
			return subs
		}
		return aggregateIPv6(subs)
	default:
		return nil
	}
}

// Summarize summarizes the given address range and returns a list of
// address prefixes.
func Summarize(first, last net.IP) []Prefix {
	if first == nil || last == nil {
		return nil
	}
	if firstip := first.To4(); firstip != nil {
		lastip := last.To4()
		if lastip == nil {
			return nil
		}
		return summarizeIPv4(firstip, lastip)
	} else if firstip := first.To16(); firstip != nil && firstip.To4() == nil {
		lastip := last.To16()
		if lastip == nil || last.To4() != nil {
			return nil
		}
		return summarizeIPv6(firstip, lastip)
	} else {
		return nil
	}
}

func sortAndDedup(ps []Prefix) []Prefix {
	switch ps[0].(type) {
	case *IPv4:
		ps = byAddrFamily(ps).ipv4Only()
		sort.Sort(byAddrAndLen(ps))
	case *IPv6:
		ps = byAddrFamily(ps).ipv6Only()
		sort.Sort(byAddrAndLen(ps))
	default:
		return nil
	}
	nps := ps[:0]
	var prev Prefix
	for _, p := range ps {
		if prev == nil {
			nps = append(nps, p)
		} else if !prev.Equal(p) {
			nps = append(nps, p)
		}
		prev = p
	}
	return nps
}

type byAddrFamily []Prefix

func (ps byAddrFamily) ipv4Only() []Prefix {
	nps := ps[:0]
	for _, p := range ps {
		if p, ok := p.(*IPv4); ok {
			nps = append(nps, p)
		}
	}
	return nps
}

func (ps byAddrFamily) ipv6Only() []Prefix {
	nps := ps[:0]
	for _, p := range ps {
		if p, ok := p.(*IPv6); ok {
			nps = append(nps, p)
		}
	}
	return nps
}

type byAddrAndLen []Prefix

func (ps byAddrAndLen) Len() int {
	return len(ps)
}

func (ps byAddrAndLen) Less(i, j int) bool {
	if ncmp := Compare(ps[i], ps[j]); ncmp < 0 {
		return true
	} else if ncmp > 0 {
		return false
	}
	if ps[i].Len() < ps[j].Len() {
		return true
	} else if ps[i].Len() > ps[j].Len() {
		return false
	}
	return false
}

func (ps byAddrAndLen) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}
