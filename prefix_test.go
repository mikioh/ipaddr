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

func toPrefix(s string) (ipaddr.Prefix, error) {
	if s == "" {
		return nil, nil
	}
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		return nil, err
	}
	prefixLen, _ := n.Mask.Size()
	p, err := ipaddr.NewPrefix(n.IP, prefixLen)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func toPrefixes(ss []string) ([]ipaddr.Prefix, error) {
	var nn []*net.IPNet
	for _, s := range ss {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			return nil, err
		}
		nn = append(nn, n)
	}
	var ps []ipaddr.Prefix
	for _, n := range nn {
		prefixLen, _ := n.Mask.Size()
		p, err := ipaddr.NewPrefix(n.IP, prefixLen)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	return ps, nil
}

var containsTests = []struct {
	in  string
	sub net.IP
	ok  bool
}{
	{"192.168.0.0/24", net.ParseIP("192.168.0.1"), true},

	{"192.168.0.0/24", net.ParseIP("192.168.1.1"), false},

	{"2001:db8:f001::/48", net.ParseIP("2001:db8:f001::1"), true},

	{"2001:db8:f001::/48", net.ParseIP("2001:db8:f002::1"), false},
}

func TestContains(t *testing.T) {
	for i, tt := range containsTests {
		p, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if ok := p.Contains(tt.sub); ok != tt.ok {
			t.Errorf("#%v: got %v; want %v", i, ok, tt.ok)
		}
	}
}

var overlapsTests = []struct {
	in     string
	others []string
	ok     bool
}{
	{"192.168.0.0/24", []string{"192.168.0.0/25", "192.168.0.64/26"}, true},

	{"192.168.0.0/26", []string{"192.168.1.112/28", "192.168.1.128/28"}, false},

	{"2001:db8:f001::/48", []string{"2001:db8:f001:4000::/49", "2001:db8:f001:8000::/49"}, true},

	{"2001:db8:f001::/48", []string{"2001:db8:f002:4000::/49", "2001:db8:f002:8000::/49"}, false},
}

func TestOverlaps(t *testing.T) {
	for i, tt := range overlapsTests {
		p1, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		others, err := toPrefixes(tt.others)
		if err != nil {
			t.Fatal(err)
		}
		p2 := ipaddr.Supernet(others)
		if ok := p1.Overlaps(p2); ok != tt.ok {
			t.Errorf("#%v: got %v; want %v", i, ok, tt.ok)
		}
		if ok := p2.Overlaps(p1); ok != tt.ok {
			t.Errorf("#%v: got %v; want %v", i, ok, tt.ok)
		}
	}
}

var numAddrTests = []struct {
	in     string
	naddrs *big.Int
}{
	{"192.168.0.0/0", big.NewInt(1 << 32)},
	{"192.168.0.0/16", big.NewInt(1 << 16)},
	{"192.168.0.0/32", big.NewInt(1)},

	{"2001:db8::/0", new(big.Int).Exp(big.NewInt(2), big.NewInt(128), nil)},
	{"2001:db8::/32", new(big.Int).Exp(big.NewInt(2), big.NewInt(96), nil)},
	{"2001:db8::/64", new(big.Int).Exp(big.NewInt(2), big.NewInt(64), nil)},
	{"2001:db8::/96", new(big.Int).Exp(big.NewInt(2), big.NewInt(32), nil)},
	{"2001:db8::/128", new(big.Int).Exp(big.NewInt(2), big.NewInt(0), nil)},
}

func TestNumAddr(t *testing.T) {
	for _, tt := range numAddrTests {
		p, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if p.NumAddr().String() != tt.naddrs.String() {
			t.Errorf("%v: got %v; want %v", p, p.NumAddr().String(), tt.naddrs.String())
		}
	}
}

var bitsTests = []struct {
	in         string
	pos, nbits int
	bits       uint32
}{
	{"192.168.0.0/24", 8, 3, 0x5},

	{"2001:db8::cafe/127", 111, 3, 0x3},
}

func TestBits(t *testing.T) {
	for i, tt := range bitsTests {
		p, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if bits := p.Bits(tt.pos, tt.nbits); bits != tt.bits {
			t.Errorf("#%v: got %x; want %x", i, bits, tt.bits)
		}
	}
}

var addrTests = []struct {
	in        string
	prefixLen int
	ip, last  net.IP
}{
	{"192.168.255.255/16", 16, net.ParseIP("192.168.0.0"), net.ParseIP("192.168.255.255")},
	{"192.168.0.255/24", 24, net.ParseIP("192.168.0.0"), net.ParseIP("192.168.0.255")},

	{"2001:db8:0:0:1:2:3:cafe/64", 64, net.ParseIP("2001:db8::"), net.ParseIP("2001:db8::ffff:ffff:ffff:ffff")},
	{"2001:db8::ca7e/121", 121, net.ParseIP("2001:db8::ca00"), net.ParseIP("2001:db8::ca7f")},
}

func TestAddr(t *testing.T) {
	for _, tt := range addrTests {
		p, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if !p.Addr().Equal(tt.ip) {
			t.Errorf("%v: got %v; want %v", p, p.Addr(), tt.ip)
		} else if p.Len() != tt.prefixLen {
			t.Errorf("%v: got %v; want %v", p, p.Len(), tt.prefixLen)
		} else if !p.LastAddr().Equal(tt.last) {
			t.Errorf("%v: got %v; want %v", p, p.LastAddr(), tt.last)
		}
	}
}

var maskTests = []struct {
	in      string
	netmask net.IPMask
}{
	{"192.168.255.255/16", net.CIDRMask(16, ipaddr.IPv4PrefixLen)},

	{"2001:db8::/64", net.CIDRMask(64, ipaddr.IPv6PrefixLen)},
}

func TestMask(t *testing.T) {
	for _, tt := range maskTests {
		p, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.Compare(p.Netmask(), tt.netmask) != 0 {
			t.Errorf("%v: got %v; want %v", p, p.Netmask(), tt.netmask)
		}
		hostmask := invmask(tt.netmask)
		if bytes.Compare(p.Hostmask(), hostmask) != 0 {
			t.Errorf("%v: got %v; want %v", p, p.Hostmask(), hostmask)
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
	in     string
	begin  net.IP
	nhosts int
}{
	{"192.168.1.0/24", net.ParseIP("192.168.0.255"), 0},
	{"192.168.1.0/24", nil, 254},
	{"192.168.1.0/24", net.ParseIP("192.168.1.0"), 254},
	{"192.168.1.0/24", net.ParseIP("192.168.1.253"), 2},
	{"192.168.1.0/24", net.ParseIP("192.168.1.254"), 1},
	{"192.168.1.0/24", net.ParseIP("192.168.1.255"), 0},
	{"192.168.1.0/24", net.ParseIP("192.168.2.0"), 0},

	{"2001:db8:f001:f002::/64", net.ParseIP("2001:db8:f001:f002:ffff:ffff:ffff:fffe"), 2},
	{"2001:db8:f001:f002::/126", net.ParseIP("2001:db8:f001:f002::"), 3},
	{"2001:db8:f001:f002::/126", net.ParseIP("2001:db8:f001:f002::1"), 3},
	{"2001:db8:f001:f002::/126", net.ParseIP("2001:db8:f001:f003::"), 0},
	{"2001:db8:f001:f002::/64", net.ParseIP("2001:db8:f001:f002:ffff:ffff:ffff:ffff"), 1},
}

func TestHosts(t *testing.T) {
	for i, tt := range hostsTests {
		p, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if hosts := p.Hosts(tt.begin); len(hosts) != tt.nhosts {
			t.Errorf("#%v: got %v; want %v", i, len(hosts), tt.nhosts)
		}
	}
}

var hostIterTests = []struct {
	in          string
	first, next net.IP
}{
	{"0.0.0.0/0", nil, net.ParseIP("0.0.0.1")},
	{"0.0.0.0/0", net.ParseIP("0.0.0.0"), net.ParseIP("0.0.0.1")},
	{"0.0.0.0/0", net.ParseIP("255.255.255.254"), net.ParseIP("255.255.255.255")},
	{"0.0.0.0/0", net.ParseIP("255.255.255.255"), nil},

	{"192.168.1.0/24", net.ParseIP("192.168.0.255"), nil},
	{"192.168.1.0/24", nil, net.ParseIP("192.168.1.1")},
	{"192.168.1.0/24", net.ParseIP("192.168.1.0"), net.ParseIP("192.168.1.1")},
	{"192.168.1.0/24", net.ParseIP("192.168.1.253"), net.ParseIP("192.168.1.254")},
	{"192.168.1.0/24", net.ParseIP("192.168.1.254"), nil},
	{"192.168.1.0/24", net.ParseIP("192.168.1.255"), nil},
	{"192.168.1.0/24", net.ParseIP("192.168.2.0"), nil},

	{"::/0", nil, net.ParseIP("::1")},
	{"::/0", net.ParseIP("::"), net.ParseIP("::1")},
	{"::/0", net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe"), net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff")},
	{"::/0", net.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"), nil},

	{"2001:db8:0:1::/64", net.ParseIP("2001:db8::ffff:ffff:ffff:fffe"), nil},
	{"2001:db8:0:1::/64", nil, net.ParseIP("2001:db8:0:1::1")},
	{"2001:db8:0:1::/64", net.ParseIP("2001:db8:0:1::"), net.ParseIP("2001:db8:0:1::1")},
	{"2001:db8:0:1::/64", net.ParseIP("2001:db8:0:1:ffff:ffff:ffff:fffe"), net.ParseIP("2001:db8:0:1:ffff:ffff:ffff:ffff")},
	{"2001:db8:0:1::/64", net.ParseIP("2001:db8:0:1:ffff:ffff:ffff:ffff"), nil},
	{"2001:db8:0:1::/64", net.ParseIP("2001:db8:0:2::"), nil},
}

func TestHostIter(t *testing.T) {
	for i, tt := range hostIterTests {
		p, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if next := <-p.HostIter(tt.first); next == nil && tt.next != nil {
			t.Errorf("#%v: got %v; want %v", i, next, tt.next)
		} else if !next.Equal(tt.next) {
			t.Errorf("#%v: got %v; want %v", i, next, tt.next)
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
			t.Fatal(err)
		}
		subs := p.Subnets(tt.nbits)
		if len(subs) != 1<<uint(tt.nbits) {
			t.Errorf("%v: got %v; want %v", p, len(subs), 1<<uint(tt.nbits))
		}
		for _, s := range subs {
			if s.Len() != tt.prefixLen+tt.nbits {
				t.Errorf("%v: got %v; want %v", p, s.Len(), tt.prefixLen+tt.nbits)
			}
		}
		if super := ipaddr.Supernet(subs); super == nil {
			for _, s := range subs {
				t.Logf("subnet: %v", s)
			}
			t.Errorf("#%v: got %v; want %v", i, super, p)
		}
	}
}

var subnetIterTests = []struct {
	in    string
	nbits int
	subs  []string
}{
	{
		"192.168.1.0/29", 2,
		[]string{
			"192.168.1.0/31",
			"192.168.1.2/31",
			"192.168.1.4/31",
			"192.168.1.6/31",
		},
	},

	{
		"2001:db8::/64", 3,
		[]string{
			"2001:db8::/67",
			"2001:db8:0:0:2000::/67",
			"2001:db8:0:0:4000::/67",
			"2001:db8:0:0:6000::/67",
			"2001:db8:0:0:8000::/67",
			"2001:db8:0:0:a000::/67",
			"2001:db8:0:0:c000::/67",
			"2001:db8:0:0:e000::/67",
		},
	},
}

func TestSubnetIter(t *testing.T) {
	for _, tt := range subnetIterTests {
		p, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		i := 0
		for s := range p.SubnetIter(tt.nbits) {
			if s.String() != tt.subs[i] {
				t.Errorf("got %v; want %v", s, tt.subs[i])
			}
			i++
		}
		if i != len(tt.subs) {
			t.Errorf("got %v; want %v", i, len(tt.subs))
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
			t.Fatal(err)
		}
		excl, err := ipaddr.NewPrefix(tt.exclAddr, tt.exclPrefixLen)
		if err != nil {
			t.Fatal(err)
		}
		subs := p.Exclude(excl)
		if len(subs) != tt.exclPrefixLen-tt.prefixLen {
			for _, s := range subs {
				t.Logf("subnet: %v", s)
			}
			t.Errorf("#%v: got %v; want %v", i, len(subs), tt.exclPrefixLen-tt.prefixLen)
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
			t.Errorf("#%v: got %v; want %v", i, sum.String(), diff.String())
		}
	}
}

var setTests = []struct {
	in string
}{
	{"192.168.0.1/32"},

	{"2001:db8::1/128"},
}

func TestSet(t *testing.T) {
	for _, tt := range setTests {
		p, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		ip := p.Addr()
		ip[len(ip)-1]++
		if err := p.Set(ip, p.Len()-1); err != nil {
			t.Fatal(err)
		}
		ip[len(ip)-1]--
		if err := p.Set(ip, p.Len()+1); err != nil {
			t.Fatal(err)
		}
		p1, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if !p.Equal(p1) {
			t.Fatal(err)
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
			t.Fatal(err)
		}
		out, err := p1.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("got %#v; want %#v", out, tt.out)
		}
		p2, err := ipaddr.NewPrefix(tt.addr, 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := p2.UnmarshalBinary(tt.out); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(p2, p1) {
			t.Errorf("got %#v; want %#v", p2, p1)
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
			t.Fatal(err)
		}
		out, err := p1.MarshalText()
		if err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(out, tt.out) {
			t.Errorf("got %#v; want %#v", out, tt.out)
		}
		p2, err := ipaddr.NewPrefix(tt.addr, 0)
		if err != nil {
			t.Fatal(err)
		}
		if err := p2.UnmarshalText(tt.out); err != nil {
			t.Fatal(err)
		} else if !reflect.DeepEqual(p2, p1) {
			t.Errorf("got %#v; want %#v", p2, p1)
		}
	}
}

var compareTests = []struct {
	in   []string
	ncmp int
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

	{[]string{"192.168.0.1/24", "2001:db8:1::1/64"}, -1},
	{[]string{"2001:db8:1::1/64", "192.168.0.1/24"}, -1},
}

func TestCompare(t *testing.T) {
	for i, tt := range compareTests {
		ps, err := toPrefixes(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		if n := ipaddr.Compare(ps[0], ps[1]); n != tt.ncmp {
			t.Errorf("#%v: got %v; want %v", i, n, tt.ncmp)
		}
	}
}

var supernetTests = []struct {
	in  []string
	out string
}{
	{
		[]string{"192.168.0.0/24", "192.168.1.0/24"},
		"192.168.0.0/23",
	},
	{
		[]string{"192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/25"},
		"192.168.0.0/22",
	},
	{
		[]string{"118.168.101.0/27", "70.168.100.0/17", "102.168.103.0/26"},
		"64.0.0.0/2",
	},
	{
		[]string{"10.40.101.1/32", "10.40.102.1/32", "11.40.103.1/32"},
		"10.0.0.0/7",
	},
	{
		[]string{"192.168.200.0/24", "192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/24", "192.168.100.0/24"},
		"192.168.0.0/16",
	},

	{
		[]string{"128.0.0.0/24", "192.0.0.0/24", "65.0.0.0/24"},
		"",
	},
	{
		[]string{"0.0.0.0/0", "192.0.0.0/24", "65.0.0.0/24"},
		"",
	},

	{
		[]string{"2001:db8:1::/32", "2001:db8:2::/39"},
		"2001:db8::/32",
	},

	{
		[]string{"8001:db8:1::/34", "2013:db8:2::/32"},
		"",
	},

	{
		[]string{"192.168.0.1/24", "2013:db8:1::1/64", "192.168.1.1/24"},
		"192.168.0.0/23",
	},
	{
		[]string{"2013:db8:1::1/64", "192.168.0.1/24", "2013:db8:2::1/64"},
		"2013:db8::/46",
	},

	{
		[]string{"192.168.0.1/24", "2013:db8:1::1/64", "1.1.1.1/24"},
		"",
	},
	{
		[]string{"2013:db8:1::1/64", "192.168.0.1/24", "8001:db8:1::1/64"},
		"",
	},
}

func TestSupernet(t *testing.T) {
	for i, tt := range supernetTests {
		subs, err := toPrefixes(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		want, err := toPrefix(tt.out)
		if err != nil {
			t.Fatal(err)
		}
		super := ipaddr.Supernet(subs)
		if !reflect.DeepEqual(super, want) {
			t.Errorf("#%v: got %v; want %v", i, super, want)
		}
	}
}

var aggregateTests = []struct {
	in  []string
	out []string
}{
	{
		[]string{"192.168.0.0/32", "192.168.0.1/32", "192.168.0.2/32", "192.168.0.3/32", "192.168.0.4/32", "192.168.0.0/32", "192.168.0.1/32"},
		[]string{"192.168.0.0/30", "192.168.0.4/32"},
	},
	{
		[]string{"192.168.0.0/22", "192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/24", "192.168.3.0/24", "192.168.4.0/24"},
		[]string{"192.168.0.0/22", "192.168.4.0/24"},
	},
	{
		[]string{"192.168.0.0/24", "192.168.1.0/24"},
		[]string{"192.168.0.0/23"},
	},
	{
		[]string{"192.168.0.1/32", "192.168.0.1/32"},
		[]string{"192.168.0.1/32"},
	},
	{
		[]string{"192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/25"},
		[]string{"192.168.0.0/23", "192.168.2.0/25"},
	},
	{
		[]string{"65.0.0.0/24", "128.0.0.0/24", "192.0.0.0/24"},
		[]string{"65.0.0.0/24", "128.0.0.0/24", "192.0.0.0/24"},
	},
	{
		[]string{"0.0.0.0/0", "192.0.0.0/24", "65.0.0.0/24"},
		[]string{"0.0.0.0/0", "65.0.0.0/24", "192.0.0.0/24"},
	},
	{
		[]string{"192.168.100.101/32", "192.168.100.102/31", "192.168.100.104/29", "192.168.100.112/28", "192.168.100.128/25", "192.168.101.0/32"},
		[]string{"192.168.100.101/32", "192.168.100.102/31", "192.168.100.104/29", "192.168.100.112/28", "192.168.100.128/25", "192.168.101.0/32"},
	},
	{
		[]string{"0.0.0.0/0", "0.0.0.0/0", "255.255.255.255/32", "255.255.255.255/32"},
		[]string{"0.0.0.0/0", "255.255.255.255/32"},
	},

	{
		[]string{"2001:db8::/64", "2001:db8:0:1::/64", "2001:db8:0:2::/64", "2001:db8:0:3::/64", "2001:db8:0:4::/64", "2001:db8::/64", "2001:db8::1/64"},
		[]string{"2001:db8::/62", "2001:db8:0:4::/64"},
	},
	{
		[]string{"::/0", "2001:db8::/32", "ff30:1:2:3::/64"},
		[]string{"::/0", "2001:db8::/32", "ff30:1:2:3::/64"},
	},
	{
		[]string{"::/0", "::/0", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"},
		[]string{"::/0", "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff/128"},
	},
}

func TestAggregate(t *testing.T) {
	for i, tt := range aggregateTests {
		subs, err := toPrefixes(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		wants, err := toPrefixes(tt.out)
		if err != nil {
			t.Fatal(err)
		}
		aggrs := ipaddr.Aggregate(subs)
		if !reflect.DeepEqual(aggrs, wants) {
			t.Errorf("#%v: got %v; want %v", i, aggrs, wants)
		}
	}
}

var subnetsAndAggregateTests = []struct {
	in string
}{
	{"172.16.0.0/16"},

	{"2001:db8::/48"},
}

func TestSubnetsAndAggregate(t *testing.T) {
	for i, tt := range subnetsAndAggregateTests {
		want, err := toPrefix(tt.in)
		if err != nil {
			t.Fatal(err)
		}
		subs := want.Subnets(12)
		aggrs := ipaddr.Aggregate(subs)
		if !reflect.DeepEqual(aggrs, []ipaddr.Prefix{want}) {
			t.Errorf("#%v: got %v; want %v", i, aggrs, []ipaddr.Prefix{want})
		}
	}
}
