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
	p, _ := ipaddr.NewPrefix(net.ParseIP("192.168.255.0"), 24)
	ip := net.ParseIP("192.168.255.255")
	for i := 0; i < b.N; i++ {
		p.Contains(ip)
	}
}

func BenchmarkIPv4Overlaps(b *testing.B) {
	p1, _ := ipaddr.NewPrefix(net.ParseIP("192.168.1.0"), 24)
	p2, _ := ipaddr.NewPrefix(net.ParseIP("192.168.2.0"), 24)
	for i := 0; i < b.N; i++ {
		p1.Overlaps(p2)
	}
}

func BenchmarkIPv4Equal(b *testing.B) {
	p1, _ := ipaddr.NewPrefix(net.ParseIP("192.168.1.0"), 24)
	p2, _ := ipaddr.NewPrefix(net.ParseIP("192.168.2.0"), 24)
	for i := 0; i < b.N; i++ {
		p1.Equal(p2)
	}
}

func BenchmarkIPv4Subnets(b *testing.B) {
	p, _ := ipaddr.NewPrefix(net.ParseIP("192.168.0.0"), 16)
	for i := 0; i < b.N; i++ {
		p.Subnets(3)
	}
}

func BenchmarkIPv4Exclude(b *testing.B) {
	p, _ := ipaddr.NewPrefix(net.ParseIP("10.1.0.0"), 16)
	p1, _ := ipaddr.NewPrefix(net.ParseIP("10.1.1.1"), 32)
	for i := 0; i < b.N; i++ {
		p.Exclude(p1)
	}
}

func BenchmarkIPv4ComparePrefix(b *testing.B) {
	p1, _ := ipaddr.NewPrefix(net.ParseIP("192.168.1.0"), 24)
	p2, _ := ipaddr.NewPrefix(net.ParseIP("192.168.2.0"), 24)
	for i := 0; i < b.N; i++ {
		ipaddr.ComparePrefix(p1, p2)
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
		p, _ := ipaddr.NewPrefix(n.IP, l)
		subs = append(subs, p)
	}
	for i := 0; i < b.N; i++ {
		ipaddr.SummaryPrefix(subs)
	}
}
