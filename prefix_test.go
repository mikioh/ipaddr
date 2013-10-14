// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr_test

import (
	"bytes"
	"math/big"
	"net"
	"reflect"
	"testing"

	"github.com/mikioh/ipaddr"
)

var containsTests = []struct {
	addr      net.IP
	prefixLen int
	sub       net.IP
	ok        bool
}{
	{net.ParseIP("192.168.0.0"), 24, net.ParseIP("192.168.0.1"), true},

	{net.ParseIP("192.168.0.0"), 24, net.ParseIP("192.168.1.1"), false},

	{net.ParseIP("2001:db8:f001::"), 48, net.ParseIP("2001:db8:f001::1"), true},

	{net.ParseIP("2001:db8:f001::"), 48, net.ParseIP("2001:db8:f002::1"), false},
}

func TestContains(t *testing.T) {
	for i, tt := range containsTests {
		if p, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen); err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		} else if ok := p.Contains(tt.sub); ok != tt.ok {
			t.Fatalf("#%v: got %v; expected %v", i, ok, tt.ok)
		}
	}
}

var overlapsTests = []struct {
	addr      net.IP
	prefixLen int
	others    []string
	ok        bool
}{
	{net.ParseIP("192.168.0.0"), 24, []string{"192.168.0.0/25", "192.168.0.64/26"}, true},

	{net.ParseIP("192.168.0.0"), 26, []string{"192.168.1.112/28", "192.168.1.128/28"}, false},

	{net.ParseIP("2001:db8:f001::"), 48, []string{"2001:db8:f001:4000::/49", "2001:db8:f001:8000::/49"}, true},

	{net.ParseIP("2001:db8:f001::"), 48, []string{"2001:db8:f002:4000::/49", "2001:db8:f002:8000::/49"}, false},
}

func TestOverlaps(t *testing.T) {
	for i, tt := range overlapsTests {
		p1, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		var nn []*net.IPNet
		for _, s := range tt.others {
			_, n, err := net.ParseCIDR(s)
			if err != nil {
				t.Fatalf("net.ParseCIDR failed: %v", err)
			}
			nn = append(nn, n)
		}
		var others []ipaddr.Prefix
		for _, n := range nn {
			prefixLen, _ := n.Mask.Size()
			p, err := ipaddr.NewPrefix(n.IP, prefixLen)
			if err != nil {
				t.Fatalf("ipaddr.NewPrefix failed: %v", err)
			}
			others = append(others, p)
		}
		p2 := ipaddr.SummaryPrefix(others)
		if ok := p1.Overlaps(p2); ok != tt.ok {
			t.Fatalf("#%v: got %v; expected %v", i, ok, tt.ok)
		}
		if ok := p2.Overlaps(p1); ok != tt.ok {
			t.Fatalf("#%v: got %v; expected %v", i, ok, tt.ok)
		}
	}
}

var numAddrTests = []struct {
	addr      net.IP
	prefixLen int
	naddrs    *big.Int
}{
	{net.ParseIP("192.168.0.0"), 0, big.NewInt(1 << 32)},
	{net.ParseIP("192.168.0.0"), 16, big.NewInt(1 << 16)},
	{net.ParseIP("192.168.0.0"), 32, big.NewInt(1)},

	{net.ParseIP("2001:db8::"), 0, new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)},
	{net.ParseIP("2001:db8::"), 32, new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil)},
	{net.ParseIP("2001:db8::"), 64, new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil)},
	{net.ParseIP("2001:db8::"), 96, new(big.Int).Exp(big.NewInt(2), big.NewInt(32), nil)},
	{net.ParseIP("2001:db8::"), 128, new(big.Int).Exp(big.NewInt(2), big.NewInt(0), nil)},
}

func TestNumAddr(t *testing.T) {
	for _, tt := range numAddrTests {
		if p, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen); err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		} else if p.NumAddr().String() != tt.naddrs.String() {
			t.Fatalf("%v: got %v; expected %v", p, p.NumAddr().String(), tt.naddrs.String())
		}
	}
}

var bitsTests = []struct {
	addr       net.IP
	prefixLen  int
	pos, nbits int
	bits       uint32
}{
	{net.ParseIP("192.168.0.0"), 24, 8, 3, 0x5},

	{net.ParseIP("2001:db8::cafe"), 127, 111, 3, 0x3},
}

func TestBits(t *testing.T) {
	for i, tt := range bitsTests {
		if p, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen); err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		} else if bits := p.Bits(tt.pos, tt.nbits); bits != tt.bits {
			t.Fatalf("#%v: got %x; expected %x", i, bits, tt.bits)
		}
	}
}

var addrTests = []struct {
	addr      net.IP
	prefixLen int
	ip, last  net.IP
}{
	{net.ParseIP("192.168.255.255"), 16, net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
	{net.ParseIP("192.168.0.255"), 24, net.ParseIP("192.168.0.0"), net.ParseIP("192.168.0.255")},

	{net.ParseIP("2001:db8:0:0:1:2:3::cafe"), 64, net.ParseIP("2001:db8::"), net.ParseIP("2001:db8::ffff:ffff:ffff:ffff")},
	{net.ParseIP("2001:db8::ca7e"), 121, net.ParseIP("2001:db8::ca00"), net.ParseIP("2001:db8::ca7f")},
}

func TestAddr(t *testing.T) {
	for _, tt := range addrTests {
		if p, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen); err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		} else if !p.Addr().Equal(tt.ip) {
			t.Fatalf("%v: got %v; expected %v", p, p.Addr(), tt.ip)
		} else if p.Len() != tt.prefixLen {
			t.Fatalf("%v: got %v; expected %v", p, p.Len(), tt.prefixLen)
		} else if !p.LastAddr().Equal(tt.last) {
			t.Fatalf("%v: got %v; expected %v", p, p.LastAddr(), tt.last)
		}
	}
}

var maskTests = []struct {
	addr      net.IP
	prefixLen int
	netmask   net.IPMask
}{
	{net.ParseIP("192.168.255.255"), 16, net.CIDRMask(16, ipaddr.IPv4PrefixLen)},

	{net.ParseIP("2001:db8::"), 64, net.CIDRMask(64, ipaddr.IPv6PrefixLen)},
}

func TestMask(t *testing.T) {
	for _, tt := range maskTests {
		p, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		} else if bytes.Compare(p.Netmask(), tt.netmask) != 0 {
			t.Fatalf("%v: got %v; expected %v", p, p.Netmask(), tt.netmask)
		}
		hostmask := invmask(tt.netmask)
		if bytes.Compare(p.Hostmask(), hostmask) != 0 {
			t.Fatalf("%v: got %v; expected %v", p, p.Hostmask(), hostmask)
		}
	}
}

func invmask(m net.IPMask) net.IPMask {
	m1 := make(net.IPMask, len(m))
	copy(m1, m)
	for i := range m {
		m1[i] = ^m[i]
	}
	return m1
}

var hostsTests = []struct {
	addr      net.IP
	prefixLen int
	begin     net.IP
	nhosts    int
}{
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.0.255"), 0},
	{net.ParseIP("192.168.1.0"), 24, nil, 254},
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.1.0"), 254},
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.1.253"), 2},
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.1.254"), 1},
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.1.255"), 0},
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.2.0"), 0},

	{net.ParseIP("2001:db8:f001:f002::"), 64, net.ParseIP("2001:db8:f001:f002::ffff:ffff:ffff:fffe"), 2},
	{net.ParseIP("2001:db8:f001:f002::"), 126, net.ParseIP("2001:db8:f001:f002::"), 3},
	{net.ParseIP("2001:db8:f001:f002::"), 126, net.ParseIP("2001:db8:f001:f002::1"), 3},
	{net.ParseIP("2001:db8:f001:f002::"), 126, net.ParseIP("2001:db8:f001:f003::"), 0},
	{net.ParseIP("2001:db8:f001:f002::"), 64, net.ParseIP("2001:db8:f001:f002::ffff:ffff:ffff:ffff"), 1},
}

func TestHosts(t *testing.T) {
	for i, tt := range hostsTests {
		p, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		if hosts := p.Hosts(tt.begin); len(hosts) != tt.nhosts {
			t.Errorf("#%v: got %v; expected %v", i, len(hosts), tt.nhosts)
		}
	}
}

var hostIterTests = []struct {
	addr        net.IP
	prefixLen   int
	first, next net.IP
}{
	{net.ParseIP("0.0.0.0"), 0, nil, net.ParseIP("0.0.0.1")},
	{net.ParseIP("0.0.0.0"), 0, net.ParseIP("0.0.0.0"), net.ParseIP("0.0.0.1")},
	{net.ParseIP("0.0.0.0"), 0, net.ParseIP("255.255.255.254"), net.ParseIP("255.255.255.255")},
	{net.ParseIP("0.0.0.0"), 0, net.ParseIP("255.255.255.255"), nil},

	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.0.255"), nil},
	{net.ParseIP("192.168.1.0"), 24, nil, net.ParseIP("192.168.1.1")},
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.1.0"), net.ParseIP("192.168.1.1")},
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.1.253"), net.ParseIP("192.168.1.254")},
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.1.254"), nil},
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.1.255"), nil},
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.2.0"), nil},

	{net.ParseIP("::"), 0, nil, net.ParseIP("::1")},
	{net.ParseIP("::"), 0, net.ParseIP("::"), net.ParseIP("::1")},
	{net.ParseIP("::"), 0, net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe"), net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")},
	{net.ParseIP("::"), 0, net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"), nil},

	{net.ParseIP("2001:db8:0:1::"), 64, net.ParseIP("2001:db8::ffff:ffff:ffff:fffe"), nil},
	{net.ParseIP("2001:db8:0:1::"), 64, nil, net.ParseIP("2001:db8:0:1::1")},
	{net.ParseIP("2001:db8:0:1::"), 64, net.ParseIP("2001:db8:0:1::"), net.ParseIP("2001:db8:0:1::1")},
	{net.ParseIP("2001:db8:0:1::"), 64, net.ParseIP("2001:db8:0:1:ffff:ffff:ffff:fffe"), net.ParseIP("2001:db8:0:1:ffff:ffff:ffff:ffff")},
	{net.ParseIP("2001:db8:0:1::"), 64, net.ParseIP("2001:db8:0:1:ffff:ffff:ffff:ffff"), nil},
	{net.ParseIP("2001:db8:0:1::"), 64, net.ParseIP("2001:db8:0:2::"), nil},
}

func TestHostIter(t *testing.T) {
	for i, tt := range hostIterTests {
		p, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		if next := <-p.HostIter(tt.first); next == nil && tt.next != nil {
			t.Errorf("#%v: got %v; expected %v", i, next, tt.next)
		} else if !next.Equal(tt.next) {
			t.Errorf("#%v: got %v; expected %v", i, next, tt.next)
		}
	}
}

var subnetsTests = []struct {
	addr      net.IP
	prefixLen int
	nbits     int
}{
	{net.ParseIP("0.0.0.0"), 29, 2},
	{net.ParseIP("192.168.254.128"), 25, 4},

	{net.ParseIP("2001:db8::"), 65, 8},
	{net.ParseIP("2001:db8::"), 51, 9},
	{net.ParseIP("2001:db8::"), 32, 1},
	{net.ParseIP("2001:db8::"), 13, 7},
	{net.ParseIP("2001:db8::"), 64, 3},
	{net.ParseIP("2001:db8::"), 61, 5},
	{net.ParseIP("2001:db8::80"), 121, 6},
}

func TestSubnets(t *testing.T) {
	for i, tt := range subnetsTests {
		p, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		subs := p.Subnets(tt.nbits)
		if len(subs) != 1<<uint(tt.nbits) {
			t.Fatalf("%v: got %v; expected %v", p, len(subs), 1<<uint(tt.nbits))
		}
		for _, s := range subs {
			if s.Len() != tt.prefixLen+tt.nbits {
				t.Errorf("%v: got %v; expected %v", p, s.Len(), tt.prefixLen+tt.nbits)
			}
		}
		if sum := ipaddr.SummaryPrefix(subs); sum == nil {
			for _, s := range subs {
				t.Logf("subnet: %v", s)
			}
			t.Fatalf("#%v: got %v; expected %v", i, sum, p)
		}
	}
}

var subnetIterTests = []struct {
	addr      net.IP
	prefixLen int
	nbits     int
	subs      []string
}{
	{net.ParseIP("192.168.1.0"), 29, 2, []string{
		"192.168.1.0/31",
		"192.168.1.2/31",
		"192.168.1.4/31",
		"192.168.1.6/31",
	}},

	{net.ParseIP("2001:db8::"), 64, 3, []string{
		"2001:db8::/67",
		"2001:db8:0:0:2000::/67",
		"2001:db8:0:0:4000::/67",
		"2001:db8:0:0:6000::/67",
		"2001:db8:0:0:8000::/67",
		"2001:db8:0:0:a000::/67",
		"2001:db8:0:0:c000::/67",
		"2001:db8:0:0:e000::/67",
	}},
}

func TestSubnetIter(t *testing.T) {
	for _, tt := range subnetIterTests {
		p, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		i := 0
		for s := range p.SubnetIter(tt.nbits) {
			if s.String() != tt.subs[i] {
				t.Errorf("got %v; expected %v", s, tt.subs[i])
			}
			i++
		}
		if i != len(tt.subs) {
			t.Fatalf("got %v; expected %v", i, len(tt.subs))
		}
	}
}

var excludeTests = []struct {
	addr          net.IP
	prefixLen     int
	exclAddr      net.IP
	exclPrefixLen int
}{
	{net.ParseIP("10.0.0.0"), 8, net.ParseIP("10.1.0.0"), 16},
	{net.ParseIP("192.168.1.0"), 24, net.ParseIP("192.168.1.0"), 26},

	{net.ParseIP("2001:db8:f001::"), 48, net.ParseIP("2001:db8:f001:f002::"), 56},
	{net.ParseIP("2001:db8:f001:f002::"), 64, net.ParseIP("2001:db8:f001:f002::cafe"), 128},
}

func TestExclude(t *testing.T) {
	for i, tt := range excludeTests {
		p, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		excl, err := ipaddr.NewPrefix(tt.exclAddr, tt.exclPrefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		subs := p.Exclude(excl)
		if len(subs) != tt.exclPrefixLen-tt.prefixLen {
			for _, s := range subs {
				t.Logf("subnet: %v", s)
			}
			t.Fatalf("#%v: got %v; expected %v", i, len(subs), tt.exclPrefixLen-tt.prefixLen)
		}
		diff, sum := big.NewInt(0), big.NewInt(0)
		diff.Sub(p.NumAddr(), excl.NumAddr())
		for _, p := range subs {
			sum.Add(sum, p.NumAddr())
		}
		if diff.String() != sum.String() {
			for _, s := range subs {
				t.Logf("subnet: %v", s)
			}
			t.Fatalf("#%v: got %v; expected %v", i, sum.String(), diff.String())
		}
	}
}

var setTests = []struct {
	addr      net.IP
	prefixLen int
}{
	{net.ParseIP("192.168.0.1"), 32},

	{net.ParseIP("2001:db8::1"), 128},
}

func TestSet(t *testing.T) {
	for _, tt := range setTests {
		p, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		ip := p.Addr()
		ip[len(ip)-1]++
		if err := p.Set(ip, p.Len()-1); err != nil {
			t.Fatalf("ipaddr.Set failed: %v", err)
		}
		ip[len(ip)-1]--
		if err := p.Set(ip, p.Len()+1); err != nil {
			t.Fatalf("ipaddr.Set failed: %v", err)
		}
		p1, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		if !p.Equal(p1) {
			t.Fatalf("got %v; expected %v", p, p1)
		}
	}
}

var binaryMarshalerUnmarshalerTests = []struct {
	addr      net.IP
	prefixLen int
	out       []byte
}{
	{net.ParseIP("0.0.0.0"), 0, []byte{0}},
	{net.ParseIP("192.0.0.0"), 7, []byte{7, 192}},
	{net.ParseIP("192.168.0.0"), 23, []byte{23, 192, 168, 0}},

	{net.ParseIP("::"), 0, []byte{0}},
	{net.ParseIP("2001::"), 8, []byte{8, 0x20}},
	{net.ParseIP("2001:db8:0:cafe:babe::"), 66, []byte{66, 0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0xca, 0xfe, 0x80}},
	{net.ParseIP("2001:db8:0:cafe:babe::3"), 127, []byte{127, 0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0xca, 0xfe, 0xba, 0xbe, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}},
}

func TestBinaryMarshalerUnmarshaler(t *testing.T) {
	for _, tt := range binaryMarshalerUnmarshalerTests {
		p1, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		out, err := p1.MarshalBinary()
		if err != nil {
			t.Fatalf("ipaddr.Prefix.MarshalBinary failed: %v", err)
		} else if !reflect.DeepEqual(out, tt.out) {
			t.Fatalf("got %#v; expected %#v", out, tt.out)
		}
		p2, err := ipaddr.NewPrefix(tt.addr, 0)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		if err := p2.UnmarshalBinary(tt.out); err != nil {
			t.Fatalf("ipaddr.Prefix.UnmarshalBinary failed: %v", err)
		} else if !reflect.DeepEqual(p2, p1) {
			t.Fatalf("got %#v; expected %#v", p2, p1)
		}
	}
}

var textMarshalerUnmarshalerTests = []struct {
	addr      net.IP
	prefixLen int
	out       []byte
}{
	{net.ParseIP("192.168.0.0"), 24, []byte("192.168.0.0/24")},

	{net.ParseIP("2001:db8::cafe"), 127, []byte("2001:db8::cafe/127")},
}

func TestTextMarshalerUnmarshaler(t *testing.T) {
	for _, tt := range textMarshalerUnmarshalerTests {
		p1, err := ipaddr.NewPrefix(tt.addr, tt.prefixLen)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		out, err := p1.MarshalText()
		if err != nil {
			t.Fatalf("ipaddr.Prefix.MarshalText failed: %v", err)
		} else if !reflect.DeepEqual(out, tt.out) {
			t.Fatalf("got %#v; expected %#v", out, tt.out)
		}
		p2, err := ipaddr.NewPrefix(tt.addr, 0)
		if err != nil {
			t.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		if err := p2.UnmarshalText(tt.out); err != nil {
			t.Fatalf("ipaddr.Prefix.UnmarshalText failed: %v", err)
		} else if !reflect.DeepEqual(p2, p1) {
			t.Fatalf("got %#v; expected %#v", p2, p1)
		}
	}
}

var comparePrefixTests = []struct {
	prefixes []string
	ncmp     int
}{
	{[]string{"192.168.1.0/23", "192.168.1.0/24"}, -1},
	{[]string{"192.168.1.0/24", "192.168.1.0/24"}, 0},
	{[]string{"192.168.1.0/25", "192.168.1.0/24"}, +1},

	{[]string{"192.168.0.0/24", "192.168.1.0/24"}, -1},
	{[]string{"192.168.1.0/24", "192.168.1.0/24"}, 0},
	{[]string{"192.168.2.0/24", "192.168.1.0/24"}, +1},

	{[]string{"2001:db8:1::/47", "2001:db8:1::/48"}, -1},
	{[]string{"2001:db8:1::/48", "2001:db8:1::/48"}, 0},
	{[]string{"2001:db8:1::/49", "2001:db8:1::/48"}, +1},

	{[]string{"2001:db8:1::/128", "2001:db8:1::1/128"}, -1},
	{[]string{"2001:db8:1::1/128", "2001:db8:1::1/128"}, 0},
	{[]string{"2001:db8:1::2/128", "2001:db8:1::1/128"}, +1},
}

func TestComparePrefix(t *testing.T) {
	for i, tt := range comparePrefixTests {
		var nn []*net.IPNet
		for _, s := range tt.prefixes {
			_, n, err := net.ParseCIDR(s)
			if err != nil {
				t.Fatalf("net.ParseCIDR failed: %v", err)
			}
			nn = append(nn, n)
		}
		var ps []ipaddr.Prefix
		for _, n := range nn {
			prefixLen, _ := n.Mask.Size()
			p, err := ipaddr.NewPrefix(n.IP, prefixLen)
			if err != nil {
				t.Fatalf("ipaddr.NewPrefix failed: %v", err)
			}
			ps = append(ps, p)
		}
		if n := ipaddr.ComparePrefix(ps[0], ps[1]); n != tt.ncmp {
			t.Fatalf("#%v: got %v; expected %v", i, n, tt.ncmp)
		}
	}
}

var summaryPrefixTests = []struct {
	subs []string
	ok   bool
}{
	{[]string{"1.1.0.0/24", "1.1.1.0/24"}, true},
	{[]string{"1.1.0.0/24", "1.1.1.0/24", "1.1.2.0/25"}, true},
	{[]string{"118.168.101.0/27", "70.168.100.0/17", "102.168.103.0/26"}, true},
	{[]string{"10.40.101.1/32", "10.40.102.1/32", "11.40.103.1/32"}, true},

	{[]string{"128.0.0.0/24", "192.0.0.0/24", "65.0.0.0/24"}, false},
	{[]string{"0.0.0.0/0", "192.0.0.0/24", "65.0.0.0/24"}, false},

	{[]string{"2001:db8:1::/32", "2001:db8:2::/39"}, true},

	{[]string{"8001:db8:1::/34", "2013:db8:2::/32"}, false},
}

func TestSummaryPrefix(t *testing.T) {
	for i, tt := range summaryPrefixTests {
		var nn []*net.IPNet
		for _, s := range tt.subs {
			_, n, err := net.ParseCIDR(s)
			if err != nil {
				t.Fatalf("net.ParseCIDR failed: %v", err)
			}
			nn = append(nn, n)
		}
		var subs []ipaddr.Prefix
		for _, n := range nn {
			prefixLen, _ := n.Mask.Size()
			p, err := ipaddr.NewPrefix(n.IP, prefixLen)
			if err != nil {
				t.Fatalf("ipaddr.NewPrefix failed: %v", err)
			}
			subs = append(subs, p)
		}
		sum := ipaddr.SummaryPrefix(subs)
		if sum == nil && tt.ok || sum != nil && !tt.ok {
			t.Fatalf("#%v: got %v, %v; expected %v", i, sum, sum != nil, tt.ok)
		}
		if tt.ok {
			for _, s := range subs {
				if !sum.Contains(s.Addr()) {
					t.Fatalf("#%v: %v does not contain %v", i, sum, s)
				}
			}
		}
	}
}
