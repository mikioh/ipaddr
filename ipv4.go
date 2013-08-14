// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr

import (
	"encoding"
	"fmt"
	"math/big"
	"net"
	"time"
)

var (
	_ Prefix                   = &IPv4{}
	_ encoding.TextMarshaler   = &IPv4{}
	_ encoding.TextUnmarshaler = &IPv4{}
	_ fmt.Stringer             = &IPv4{}
)

// Maximum length of IPv4 address prefix in bits.
const IPv4PrefixLen = 8 * net.IPv4len

// An IPv4 represents an IPv4 address prefix.
type IPv4 struct {
	nbits byte    // prefix length
	addr  ipv4Int // address
}

func (p *IPv4) isDefaultRoute() bool {
	return p.addr == 0 && p.nbits == 0
}

func (p *IPv4) isNetworkAddr(i ipv4Int) bool {
	if p.nbits == IPv4PrefixLen {
		return false
	}
	return p.addr == i
}

func (p *IPv4) isBroadcastAddr(i ipv4Int) bool {
	return p.addr|ipv4Int(^mask32(p.nbits)) == i
}

func (p *IPv4) isLimitedBroadcastAddr(i ipv4Int) bool {
	return i == ipv4Int(0xffffffff)
}

func (p *IPv4) isHostAssignable(i ipv4Int) (ok bool, endOfRange bool) {
	if p.isLimitedBroadcastAddr(i) {
		if p.isDefaultRoute() {
			return true, true
		}
		return false, true
	} else if p.isBroadcastAddr(i) {
		return false, true
	} else if p.isNetworkAddr(i) {
		return false, false
	}
	return true, false
}

func (p *IPv4) contains(i ipv4Int) bool {
	return p.addr == i&ipv4Int(mask32(p.nbits))
}

// Contains implements the Contains method of ipaddr.Prefix interface.
func (p *IPv4) Contains(ip net.IP) bool {
	return p.contains(ipToIPv4Int(ip.To4()))
}

// Overlaps implements the Overlaps method of ipaddr.Prefix interface.
func (p *IPv4) Overlaps(prefix Prefix) bool {
	q, ok := prefix.(*IPv4)
	if !ok {
		return false
	}
	return p.contains(q.addr) || p.contains(q.lastAddr()) || q.contains(p.addr) || q.contains(p.lastAddr())
}

// Equal implements the Equal method of ipaddr.Prefix interface.
func (p *IPv4) Equal(prefix Prefix) bool {
	switch prefix.(type) {
	case *IPv4:
		return p.addr == prefix.(*IPv4).addr && p.nbits == prefix.(*IPv4).nbits
	}
	return false
}

// String implements the String method of fmt.Stringer interface.
func (p *IPv4) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s/%d", p.addr.IP().String(), p.nbits)
}

// Len implements the Len method of ipaddr.Prefix interface.
func (p *IPv4) Len() int {
	return int(p.nbits)
}

// NumAddr implements the NumAddr method of ipaddr.Prefix interface.
func (p *IPv4) NumAddr() *big.Int {
	i := new(big.Int).SetBytes(p.Hostmask())
	return i.Add(i, big.NewInt(1))
}

// Bits implements the Bits method of ipaddr.Prefix interface.
func (p *IPv4) Bits(pos, nbits int) uint32 {
	if 0 > pos || pos > IPv4PrefixLen-1 || 0 > nbits || nbits > 32 {
		return 0
	}
	return uint32((p.addr << uint(pos)) >> uint(IPv4PrefixLen-nbits))
}

// Addr implements the Addr method of ipaddr.Prefix interface.
func (p *IPv4) Addr() net.IP {
	return p.addr.IP()
}

func (p *IPv4) lastAddr() ipv4Int {
	return p.addr | ipv4Int(^mask32(p.nbits))
}

// LastAddr implements the LastAddr method of ipaddr.Prefix interface.
func (p *IPv4) LastAddr() net.IP {
	i := p.lastAddr()
	return i.IP()
}

// BroadcastAddr returns the directed broadcast address for the
// prefix.
func (p *IPv4) BroadcastAddr() net.IP {
	if p.nbits == IPv4PrefixLen {
		return nil
	}
	i := p.addr | ipv4Int(^mask32(p.nbits))
	return i.IP()
}

// Hostmask implements the Hostmask method of ipaddr.Prefix interface.
func (p *IPv4) Hostmask() net.IPMask {
	i := ipv4Int(^mask32(p.nbits))
	return net.IPMask(i.IP())
}

// Netmask implements the Netmask method of ipaddr.Prefix interface.
func (p *IPv4) Netmask() net.IPMask {
	return net.CIDRMask(int(p.nbits), IPv4PrefixLen)
}

// Hosts implements the Hosts method of ipaddr.Prefix interface.
func (p *IPv4) Hosts(begin net.IP) []net.IP {
	if p.isDefaultRoute() {
		return nil
	}
	var cur ipv4Int
	if len(begin) != 0 {
		cur = ipToIPv4Int(begin.To4())
	} else {
		cur = p.addr
	}
	var hosts []net.IP
	if ok, _ := p.isHostAssignable(cur); ok && p.contains(cur) {
		hosts = append(hosts, cur.IP())
	}
	if IPv4PrefixLen-p.nbits < 17 { // don't bother runtime.makeslice by big number
		for p.contains(cur) {
			if _, eor := p.isHostAssignable(cur); eor {
				break
			}
			cur++
			if ok, _ := p.isHostAssignable(cur); ok {
				hosts = append(hosts, cur.IP())
			}
		}
		return hosts
	}
	for h := range p.HostIter(begin) {
		hosts = append(hosts, h)
	}
	return hosts
}

// HostIter implements the HostIter method of ipaddr.Prefix interface.
func (p *IPv4) HostIter(first net.IP) <-chan net.IP {
	iter := &ipv4HostIter{
		p:  IPv4{addr: p.addr, nbits: p.nbits},
		ch: make(chan net.IP, 1),
	}
	if len(first) != 0 {
		iter.cur = ipToIPv4Int(first.To4())
	} else {
		iter.cur = p.addr
	}
	go iter.run()
	return iter.ch
}

// Subnets implements the Subnets method of ipaddr.Prefix interface.
func (p *IPv4) Subnets(nbits int) []Prefix {
	if nbits < 0 || p.nbits+byte(nbits) > IPv4PrefixLen {
		return nil
	}
	var subs []Prefix
	if nbits < 17 { // don't bother runtime.makeslice by big number
		subs = make([]Prefix, 1<<uint(nbits))
		off := uint(IPv4PrefixLen - p.nbits - byte(nbits))
		for i := range subs {
			subs[i] = newIPv4(p.addr|ipv4Int(i<<off), p.nbits+byte(nbits))
		}
		return subs
	}
	for s := range p.SubnetIter(nbits) {
		subs = append(subs, s)
	}
	return subs
}

// SubnetIter implements the SubnetIter method of ipaddr.Prefix
// interface.
func (p *IPv4) SubnetIter(nbits int) <-chan Prefix {
	iter := &ipv4SubnetIter{
		p:     IPv4{addr: p.addr, nbits: p.nbits},
		nbits: byte(nbits),
		cur:   p.addr,
		ch:    make(chan Prefix, 1),
	}
	go iter.run()
	return iter.ch
}

func (p *IPv4) chopup() (IPv4, IPv4) {
	return IPv4{addr: p.addr, nbits: p.nbits + 1}, IPv4{addr: p.addr | ipv4Int(1<<uint(IPv4PrefixLen-p.nbits-1)), nbits: p.nbits + 1}
}

// Exclude implements the Exclude method of ipaddr.Prefix interface.
func (p *IPv4) Exclude(prefix Prefix) []Prefix {
	var x IPv4
	switch prefix.(type) {
	case *IPv4:
		x.addr = prefix.(*IPv4).addr
		x.nbits = prefix.(*IPv4).nbits
		if !p.contains(x.addr) {
			return nil
		}
		if p.Equal(&x) {
			return []Prefix{&x}
		}
	default:
		return nil
	}
	var subs []Prefix
	l, r := p.chopup()
	for !l.Equal(&x) && !r.Equal(&x) {
		if l.contains(x.addr) {
			subs = append(subs, newIPv4(r.addr, r.nbits))
			l, r = l.chopup()
		} else if r.contains(x.addr) {
			subs = append(subs, newIPv4(l.addr, l.nbits))
			l, r = r.chopup()
		} else {
			panic("got lost in the ipv4 forest")
		}
	}
	if l.Equal(&x) {
		subs = append(subs, newIPv4(r.addr, r.nbits))
	} else if r.Equal(&x) {
		subs = append(subs, newIPv4(l.addr, l.nbits))
	}
	return subs
}

func (p *IPv4) set(i ipv4Int, nbits byte) {
	p.addr, p.nbits = i&ipv4Int(mask32(nbits)), nbits
}

// Set implements the Set method of ipaddr.Prefix interface.
func (p *IPv4) Set(ip net.IP, nbits int) error {
	if ipv4 := ip.To4(); ipv4 != nil && 0 <= nbits && nbits <= IPv4PrefixLen {
		p.set(ipToIPv4Int(ipv4), byte(nbits))
		return nil
	}
	return errInvalidArgument
}

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

func newIPv4(i ipv4Int, nbits byte) *IPv4 {
	p := &IPv4{}
	p.set(i, nbits)
	return p
}

type ipv4HostIter struct {
	p   IPv4
	cur ipv4Int
	ch  chan net.IP
}

func (iter *ipv4HostIter) run() {
	defer close(iter.ch)
	var idleTimeout <-chan time.Time
loop:
	for iter.p.contains(iter.cur) {
		if _, eor := iter.p.isHostAssignable(iter.cur); eor {
			break
		}
		iter.cur++
		if ok, _ := iter.p.isHostAssignable(iter.cur); !ok {
			continue
		}
		idleTimeout = time.After(1 * time.Second)
		select {
		case <-idleTimeout:
			break loop
		case iter.ch <- iter.cur.IP():
		}
	}
}

type ipv4SubnetIter struct {
	p      IPv4
	nbits  byte
	cur    ipv4Int
	passed bool
	ch     chan Prefix
}

func (iter *ipv4SubnetIter) run() {
	defer close(iter.ch)
	if iter.nbits < 0 || iter.p.nbits+iter.nbits > IPv4PrefixLen {
		return
	}
	m := ipv4Int(^mask32(iter.p.nbits + iter.nbits))
	nbits := iter.p.nbits + iter.nbits
	var idleTimeout <-chan time.Time
loop:
	for !iter.p.isLimitedBroadcastAddr(iter.cur) && !iter.p.isBroadcastAddr(iter.cur|m) {
		if !iter.passed {
			iter.passed = true
		} else {
			iter.cur |= m
			iter.cur++
		}
		idleTimeout = time.After(1 * time.Second)
		select {
		case <-idleTimeout:
			break loop
		case iter.ch <- newIPv4(iter.cur, nbits):
		}
	}
}

func ipv4ComparePrefix(a, b *IPv4) int {
	if a.addr < b.addr {
		return -1
	} else if a.addr > b.addr {
		return +1
	}
	if a.nbits < b.nbits {
		return -1
	} else if a.nbits > b.nbits {
		return +1
	}
	return 0
}

func ipv4SummaryPrefix(subs []Prefix) *IPv4 {
	m := ipv4Int(mask32(subs[0].(*IPv4).nbits))
	base := subs[0].(*IPv4).addr & m
	nbits := subs[0].(*IPv4).nbits
	for _, s := range subs[1:] {
		if diff := uint32(base ^ s.(*IPv4).addr&m); diff != 0 {
			if l := nlz32(diff); l < nbits {
				nbits = l
			}
		}
	}
	if nbits == 0 {
		return nil
	}
	return newIPv4(subs[0].(*IPv4).addr, nbits)
}
