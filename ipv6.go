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
	_ Prefix                     = &IPv6{}
	_ fmt.Stringer               = &IPv6{}
	_ encoding.BinaryMarshaler   = &IPv6{}
	_ encoding.BinaryUnmarshaler = &IPv6{}
	_ encoding.TextMarshaler     = &IPv6{}
	_ encoding.TextUnmarshaler   = &IPv6{}
)

// Maximum length of IPv6 address prefix in bits.
const IPv6PrefixLen = 8 * net.IPv6len

// An IPv6 represents an IPv6 address prefix.
type IPv6 struct {
	nbits int     // prefix length
	addr  ipv6Int // address
}

func (p *IPv6) isDefaultRoute() bool {
	return p.addr[0] == 0 && p.addr[1] == 0 && p.nbits == 0
}

func (p *IPv6) isSubnetRouterAnycastAddr(i ipv6Int) bool {
	if p.nbits == IPv6PrefixLen {
		return false
	}
	return p.addr == i
}

func (p *IPv6) isLastAddr(i ipv6Int) bool {
	if p.nbits > 64 {
		return p.addr[0]|^mask64(64) == i[0] && p.addr[1]|^mask64(p.nbits-64) == i[1]
	}
	return p.addr[0]|^mask64(p.nbits) == i[0] && mask64(64) == i[1]
}

func (p *IPv6) isHostAssignable(i ipv6Int) (ok bool, endOfRange bool) {
	if p.isLastAddr(i) {
		return true, true
	} else if p.isSubnetRouterAnycastAddr(i) {
		return false, false
	}
	return true, false
}

func (p *IPv6) contains(i ipv6Int) bool {
	if p.nbits > 64 {
		return p.addr[0] == i[0] && p.addr[1] == i[1]&mask64(p.nbits-64)
	}
	return p.addr[0] == i[0]&mask64(p.nbits)
}

// Contains implements the Contains method of ipaddr.Prefix interface.
func (p *IPv6) Contains(ip net.IP) bool {
	return p.contains(ipToIPv6Int(ip.To16()))
}

// Overlaps implements the Overlaps method of ipaddr.Prefix interface.
func (p *IPv6) Overlaps(prefix Prefix) bool {
	q, ok := prefix.(*IPv6)
	if !ok {
		return false
	}
	return p.contains(q.addr) || p.contains(q.lastAddr()) || q.contains(p.addr) || q.contains(p.lastAddr())
}

// Equal implements the Equal method of ipaddr.Prefix interface.
func (p *IPv6) Equal(prefix Prefix) bool {
	switch q := prefix.(type) {
	case *IPv6:
		return p.addr == q.addr && p.nbits == q.nbits
	}
	return false
}

// String implements the String method of fmt.Stringer interface.
func (p *IPv6) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%s/%d", p.addr.IP().String(), p.nbits)
}

// Len implements the Len method of ipaddr.Prefix interface.
func (p *IPv6) Len() int {
	return p.nbits
}

// NumAddr implements the NumAddr method of ipaddr.Prefix interface.
func (p *IPv6) NumAddr() *big.Int {
	i := new(big.Int).SetBytes(p.Hostmask())
	return i.Add(i, big.NewInt(1))
}

// Bits implements the Bits method of ipaddr.Prefix interface.
func (p *IPv6) Bits(pos, nbits int) uint32 {
	if 0 > pos || pos > IPv6PrefixLen-1 || 0 > nbits || nbits > 32 {
		return 0
	}
	var bits ipv6Int
	bits[0], bits[1] = p.addr[0], p.addr[1]
	bits.lshift(pos)
	bits.rshift(IPv6PrefixLen - nbits)
	return uint32(bits[1])
}

// Addr implements the Addr method of ipaddr.Prefix interface.
func (p *IPv6) Addr() net.IP {
	return p.addr.IP()
}

func (p *IPv6) lastAddr() ipv6Int {
	var i ipv6Int
	i.setHostmask(p.nbits)
	i[0], i[1] = p.addr[0]|i[0], p.addr[1]|i[1]
	return i
}

// LastAddr implements the LastAddr method of ipaddr.Prefix interface.
func (p *IPv6) LastAddr() net.IP {
	i := p.lastAddr()
	return i.IP()
}

// Hostmask implements the Hostmask method of ipaddr.Prefix interface.
func (p *IPv6) Hostmask() net.IPMask {
	var i ipv6Int
	i.setHostmask(p.nbits)
	return net.IPMask(i.IP())
}

// Netmask implements the Netmask method of ipaddr.Prefix interface.
func (p *IPv6) Netmask() net.IPMask {
	return net.CIDRMask(p.nbits, IPv6PrefixLen)
}

// Hosts implements the Hosts method of ipaddr.Prefix interface.
func (p *IPv6) Hosts(begin net.IP) []net.IP {
	if p.isDefaultRoute() {
		return nil
	}
	var cur ipv6Int
	if len(begin) != 0 {
		cur = ipToIPv6Int(begin.To16())
	} else {
		cur = p.addr
	}
	var hosts []net.IP
	if ok, _ := p.isHostAssignable(cur); ok && p.contains(cur) {
		hosts = append(hosts, cur.IP())
	}
	if IPv6PrefixLen-p.nbits < 17 { // don't bother runtime.makeslice by big number
		for p.contains(cur) {
			if _, eor := p.isHostAssignable(cur); eor {
				break
			}
			cur.incr()
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
func (p *IPv6) HostIter(first net.IP) <-chan net.IP {
	iter := &ipv6HostIter{
		p:  IPv6{addr: p.addr, nbits: p.nbits},
		ch: make(chan net.IP, 1),
	}
	if len(first) != 0 {
		iter.cur = ipToIPv6Int(first.To16())
	} else {
		iter.cur = p.addr
	}
	go iter.run()
	return iter.ch
}

// Subnets implements the Subnets method of ipaddr.Prefix interface.
func (p *IPv6) Subnets(nbits int) []Prefix {
	if nbits < 0 || p.nbits+nbits > IPv6PrefixLen {
		return nil
	}
	var subs []Prefix
	if nbits < 17 { // don't bother runtime.makeslice by big number
		subs = make([]Prefix, 1<<uint(nbits))
		off := IPv6PrefixLen - p.nbits - nbits
		for i := range subs {
			id := ipv6Int{0, uint64(i)}
			id.lshift(off)
			subs[i] = newIPv6(ipv6Int{p.addr[0] | id[0], p.addr[1] | id[1]}, p.nbits+nbits)
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
func (p *IPv6) SubnetIter(nbits int) <-chan Prefix {
	iter := &ipv6SubnetIter{
		p:     IPv6{addr: p.addr, nbits: p.nbits},
		nbits: nbits,
		cur:   p.addr,
		ch:    make(chan Prefix, 1),
	}
	go iter.run()
	return iter.ch
}

// Exclude implements the Exclude method of ipaddr.Prefix interface.
func (p *IPv6) Exclude(prefix Prefix) []Prefix {
	var x IPv6
	switch q := prefix.(type) {
	case *IPv6:
		x.addr = q.addr
		x.nbits = q.nbits
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
	l, r := p.descend(1)
	for !l.Equal(&x) && !r.Equal(&x) {
		if l.contains(x.addr) {
			subs = append(subs, newIPv6(r.addr, r.nbits))
			l, r = l.descend(1)
		} else if r.contains(x.addr) {
			subs = append(subs, newIPv6(l.addr, l.nbits))
			l, r = r.descend(1)
		} else {
			panic("got lost in the ipv6 forest")
		}
	}
	if l.Equal(&x) {
		subs = append(subs, newIPv6(r.addr, r.nbits))
	} else if r.Equal(&x) {
		subs = append(subs, newIPv6(l.addr, l.nbits))
	}
	return subs
}

func (p *IPv6) descend(nbits int) (IPv6, IPv6) {
	id := ipv6Int{0, 1}
	id.lshift(IPv6PrefixLen - p.nbits - nbits)
	return IPv6{addr: p.addr, nbits: p.nbits + nbits}, IPv6{addr: ipv6Int{p.addr[0] | id[0], p.addr[1] | id[1]}, nbits: p.nbits + nbits}
}

func (p *IPv6) set(i ipv6Int, nbits int) {
	p.addr, p.nbits = i, nbits
	if p.nbits > 64 {
		p.addr[1] = i[1] & mask64(p.nbits-64)
	} else {
		p.addr[0] = i[0] & mask64(p.nbits)
		p.addr[1] = 0
	}
}

// Set implements the Set method of ipaddr.Prefix interface.
func (p *IPv6) Set(ip net.IP, nbits int) error {
	if ipv6 := ip.To16(); ipv6 != nil && ipv6.To4() == nil && 0 <= nbits && nbits <= IPv6PrefixLen {
		p.set(ipToIPv6Int(ipv6), nbits)
		return nil
	}
	return errInvalidArgument
}

// MarshalBinary implements the MarshalBinary method of
// encoding.BinaryMarshaler interface.
func (p *IPv6) MarshalBinary() ([]byte, error) {
	var b [1 + net.IPv6len]byte
	n := p.addr.encodeNLRI(b[:], p.nbits)
	return b[:n], nil
}

// UnmarshalBinary implements the UnmarshalBinary method of
// encoding.BinaryUnmarshaler interface.
func (p *IPv6) UnmarshalBinary(data []byte) error {
	return p.decodeNLRI(data)
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

func (a *IPv6) compare(b *IPv6) int {
	if a.addr[0] < b.addr[0] {
		return -1
	} else if a.addr[0] > b.addr[0] {
		return +1
	}
	if a.addr[1] < b.addr[1] {
		return -1
	} else if a.addr[1] > b.addr[1] {
		return +1
	}
	if a.nbits < b.nbits {
		return -1
	} else if a.nbits > b.nbits {
		return +1
	}
	return 0
}

func newIPv6(i ipv6Int, nbits int) *IPv6 {
	p := &IPv6{}
	p.set(i, nbits)
	return p
}

type ipv6HostIter struct {
	p   IPv6
	cur ipv6Int
	ch  chan net.IP
}

func (iter *ipv6HostIter) run() {
	defer close(iter.ch)
	for iter.p.contains(iter.cur) {
		if _, eor := iter.p.isHostAssignable(iter.cur); eor {
			break
		}
		iter.cur.incr()
		if ok, _ := iter.p.isHostAssignable(iter.cur); !ok {
			continue
		}
		tmo := time.NewTimer(1 * time.Second)
		select {
		case <-tmo.C:
			tmo.Stop()
			return
		case iter.ch <- iter.cur.IP():
			tmo.Stop()
		}
	}
}

type ipv6SubnetIter struct {
	p      IPv6
	nbits  int
	cur    ipv6Int
	passed bool
	ch     chan Prefix
}

func (iter *ipv6SubnetIter) run() {
	defer close(iter.ch)
	if iter.nbits < 0 || iter.p.nbits+iter.nbits > IPv6PrefixLen {
		return
	}
	var m ipv6Int
	m.setHostmask(iter.p.nbits + iter.nbits)
	nbits := iter.p.nbits + iter.nbits
	for !iter.p.isLastAddr(ipv6Int{iter.cur[0] | m[0], iter.cur[1] | m[1]}) {
		if !iter.passed {
			iter.passed = true
		} else {
			iter.cur[0], iter.cur[1] = iter.cur[0]|m[0], iter.cur[1]|m[1]
			iter.cur.incr()
		}
		tmo := time.NewTimer(1 * time.Second)
		select {
		case <-tmo.C:
			tmo.Stop()
			return
		case iter.ch <- newIPv6(iter.cur, nbits):
			tmo.Stop()
		}
	}
}

func supernetIPv6(subs []Prefix) Prefix {
	var base, m, diff ipv6Int
	m.setNetmask(subs[0].(*IPv6).nbits)
	base[0], base[1] = subs[0].(*IPv6).addr[0]&m[0], subs[0].(*IPv6).addr[1]&m[1]
	nbits := subs[0].(*IPv6).nbits
	for _, s := range subs[1:] {
		diff[0], diff[1] = (base[0]^s.(*IPv6).addr[0])&m[0], (base[1]^s.(*IPv6).addr[1])&m[1]
		if diff[0] != 0 {
			if l := nlz64(diff[0]); l < nbits {
				nbits = l
			}
		} else if diff[1] != 0 {
			if l := nlz64(diff[1]); 64+l < nbits {
				nbits = 64 + l
			}
		}
	}
	if nbits == 0 {
		return nil
	}
	return newIPv6(subs[0].(*IPv6).addr, nbits)
}

func aggregateIPv6(subs []Prefix) []Prefix {
	var aggrs []Prefix
	for len(subs) > 0 {
		if subs[0].(*IPv6).nbits == 0 {
			aggrs = append(aggrs, subs[0])
			subs = subs[1:]
			continue
		}
		bf, n := ascendIPv6(subs)
		m := 1 << uint(bf)
		if n < m {
			aggrs = append(aggrs, subs[0])
			subs = subs[1:]
			continue
		}
		p := supernetIPv6(subs[:m])
		aggrs = append(aggrs, p)
		subs = subs[m:]
		m = 0
		for _, s := range subs {
			if !p.(*IPv6).contains(s.(*IPv6).addr) {
				break
			}
			m++
		}
		subs = subs[m:]
	}
	return aggrs
}

func ascendIPv6(subs []Prefix) (int, int) {
	base := subs[0].(*IPv6)
	var m ipv6Int
	m.setNetmask(base.nbits)
	var lastBF, lastN int
	for bf := 1; bf < IPv6PrefixLen; bf++ {
		n, nfull := 0, 1<<uint(bf)
		pat, max := &ipv6Int{0, 0}, &ipv6Int{0, 1}
		max.lshift(bf)
		var maggr ipv6Int
		maggr.setNetmask(base.nbits - bf)
		for ; pat.compare(max) < 0; pat.incr() {
			npat := *pat
			(&npat).lshift(IPv6PrefixLen - base.nbits)
			var aggr ipv6Int
			aggr[0], aggr[1] = base.addr[0]&maggr[0]|npat[0], base.addr[1]&maggr[1]|npat[1]
			for _, s := range subs {
				if aggr[0]^(s.(*IPv6).addr[0]&m[0]) == 0 && aggr[1]^(s.(*IPv6).addr[1]&m[1]) == 0 {
					n++
				}
			}
		}
		if n < nfull {
			break
		}
		lastBF = bf
		lastN = n
	}
	return lastBF, lastN
}
