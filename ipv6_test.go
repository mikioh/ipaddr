// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr_test

import (
	"net"
	"testing"

	"github.com/mikioh/ipaddr"
)

func BenchmarkIPv6Contains(b *testing.B) {
	p, _ := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f002::"), 64)
	ip := net.ParseIP("2001:db8:f001:f002::1")
	for i := 0; i < b.N; i++ {
		p.Contains(ip)
	}
}

func BenchmarkIPv6Overlaps(b *testing.B) {
	p1, _ := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f002::"), 64)
	p2, _ := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f003::"), 64)
	for i := 0; i < b.N; i++ {
		p1.Overlaps(p2)
	}
}

func BenchmarkIPv6Equal(b *testing.B) {
	p1, _ := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f002::"), 64)
	p2, _ := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f003::"), 64)
	for i := 0; i < b.N; i++ {
		p1.Equal(p2)
	}
}

func BenchmarkIPv6Subnets(b *testing.B) {
	p, _ := ipaddr.NewPrefix(net.ParseIP("2001:db8::"), 60)
	for i := 0; i < b.N; i++ {
		p.Subnets(3)
	}
}

func BenchmarkIPv6Exclude(b *testing.B) {
	p, _ := ipaddr.NewPrefix(net.ParseIP("2001:db8::"), 64)
	p1, _ := ipaddr.NewPrefix(net.ParseIP("2001:db8::1:1:1:1"), 128)
	for i := 0; i < b.N; i++ {
		p.Exclude(p1)
	}
}

func BenchmarkIPv6ComparePrefix(b *testing.B) {
	p1, _ := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f002::"), 64)
	p2, _ := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f003::"), 64)
	for i := 0; i < b.N; i++ {
		ipaddr.ComparePrefix(p1, p2)
	}
}

func BenchmarkIPv6SummaryPrefix(b *testing.B) {
	var nn []*net.IPNet
	for _, ns := range []string{"2001:db8:f001:a::/64", "2001:db8:f002:b::/64", "2001:db8:f003:c::/64"} {
		_, n, err := net.ParseCIDR(ns)
		if err != nil {
			b.Fatalf("net.ParseCIDR failed: %v", err)
		}
		nn = append(nn, n)
	}
	var subs []ipaddr.Prefix
	for _, n := range nn {
		l, _ := n.Mask.Size()
		p, _ := ipaddr.NewPrefix(n.IP, l)
		subs = append(subs, p)
	}
	for i := 0; i < b.N; i++ {
		ipaddr.SummaryPrefix(subs)
	}
}
