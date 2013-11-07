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
	p, err := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f002::"), 64)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	ip := net.ParseIP("2001:db8:f001:f002::1")
	for i := 0; i < b.N; i++ {
		p.Contains(ip)
	}
}

func BenchmarkIPv6Overlaps(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f002::"), 64)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f003::"), 64)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p1.Overlaps(p2)
	}
}

func BenchmarkIPv6Equal(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f002::"), 64)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f003::"), 64)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p1.Equal(p2)
	}
}

func BenchmarkIPv6Subnets(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("2001:db8::"), 60)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p.Subnets(3)
	}
}

func BenchmarkIPv6Exclude(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("2001:db8::"), 64)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("2001:db8::1:1:1:1"), 128)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p1.Exclude(p2)
	}
}

func BenchmarkIPv6MarshalBinary(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("2001:db8:0:cafe:babe::"), 66)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p.MarshalBinary()
	}
}

func BenchmarkIPv6UnmarshalBinary(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("::"), 0)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p.UnmarshalBinary([]byte{66, 0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0xca, 0xfe, 0x80})

	}
}

func BenchmarkIPv6MarshalText(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("2001:db8::cafe"), 127)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p.MarshalText()
	}
}

func BenchmarkIPv6UnmarshalText(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("::"), 0)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p.UnmarshalText([]byte("2001:db8::cafe/127"))
	}
}

func BenchmarkIPv6Compare(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f002::"), 64)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("2001:db8:f001:f003::"), 64)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		ipaddr.Compare(p1, p2)
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
		p, err := ipaddr.NewPrefix(n.IP, l)
		if err != nil {
			b.Fatalf("ipaddr.NewPrefix failed: %v", err)
		}
		subs = append(subs, p)
	}
	for i := 0; i < b.N; i++ {
		ipaddr.SummaryPrefix(subs)
	}
}
