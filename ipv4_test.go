// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr_test

import (
	"net"
	"testing"

	"github.com/mikioh/ipaddr"
)

func BenchmarkIPv4Contains(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("192.168.255.0"), 24)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	ip := net.ParseIP("192.168.255.255")
	for i := 0; i < b.N; i++ {
		p.Contains(ip)
	}
}

func BenchmarkIPv4Overlaps(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("192.168.1.0"), 24)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("192.168.2.0"), 24)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p1.Overlaps(p2)
	}
}

func BenchmarkIPv4Equal(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("192.168.1.0"), 24)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("192.168.2.0"), 24)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p1.Equal(p2)
	}
}

func BenchmarkIPv4Subnets(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("192.168.0.0"), 16)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p.Subnets(3)
	}
}

func BenchmarkIPv4Exclude(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("10.1.0.0"), 16)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("10.1.1.1"), 32)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p1.Exclude(p2)
	}
}

func BenchmarkIPv4MarshalBinary(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("192.168.0.0"), 22)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p.MarshalBinary()
	}
}

func BenchmarkIPv4UnmarshalBinary(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("0.0.0.0"), 0)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p.UnmarshalBinary([]byte{22, 192, 168, 0})
	}
}

func BenchmarkIPv4MarshalText(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("192.168.0.0"), 24)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p.MarshalText()
	}
}

func BenchmarkIPv4UnmarshalText(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("0.0.0.0"), 0)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p.UnmarshalText([]byte("192.168.0.0/24"))
	}
}

func BenchmarkIPv4Compare(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("192.168.1.0"), 24)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("192.168.2.0"), 24)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		ipaddr.Compare(p1, p2)
	}
}

func BenchmarkIPv4SummaryPrefix(b *testing.B) {
	var nn []*net.IPNet
	for _, ns := range []string{"172.16.141.0/24", "172.16.142.0/24", "172.16.143.0/24"} {
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
