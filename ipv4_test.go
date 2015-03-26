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
		b.Fatal(err)
	}
	ip := net.ParseIP("192.168.255.255")
	for i := 0; i < b.N; i++ {
		p.Contains(ip)
	}
}

func BenchmarkIPv4ContainsPrefix(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("192.168.255.0"), 24)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("192.168.255.1"), 24)
	if err != nil {
		b.Fatalf("ipaddr.NewPrefix failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		p1.ContainsPrefix(p2)
	}
}

func BenchmarkIPv4Overlaps(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("192.168.1.0"), 24)
	if err != nil {
		b.Fatal(err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("192.168.2.0"), 24)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		p1.Overlaps(p2)
	}
}

func BenchmarkIPv4Equal(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("192.168.1.0"), 24)
	if err != nil {
		b.Fatal(err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("192.168.2.0"), 24)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		p1.Equal(p2)
	}
}

func BenchmarkIPv4Subnets(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("192.168.0.0"), 16)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		p.Subnets(3)
	}
}

func BenchmarkIPv4Exclude(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("10.1.0.0"), 16)
	if err != nil {
		b.Fatal(err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("10.1.1.1"), 32)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		p1.Exclude(p2)
	}
}

func BenchmarkIPv4MarshalBinary(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("192.168.0.0"), 22)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		p.MarshalBinary()
	}
}

func BenchmarkIPv4UnmarshalBinary(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("0.0.0.0"), 0)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		p.UnmarshalBinary([]byte{22, 192, 168, 0})
	}
}

func BenchmarkIPv4MarshalText(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("192.168.0.0"), 24)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		p.MarshalText()
	}
}

func BenchmarkIPv4UnmarshalText(b *testing.B) {
	p, err := ipaddr.NewPrefix(net.ParseIP("0.0.0.0"), 0)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		p.UnmarshalText([]byte("192.168.0.0/24"))
	}
}

func BenchmarkCompareIPv4(b *testing.B) {
	p1, err := ipaddr.NewPrefix(net.ParseIP("192.168.1.0"), 24)
	if err != nil {
		b.Fatal(err)
	}
	p2, err := ipaddr.NewPrefix(net.ParseIP("192.168.2.0"), 24)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		ipaddr.Compare(p1, p2)
	}
}

func BenchmarkSupernetIPv4(b *testing.B) {
	subs, err := toPrefixes([]string{"192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/24", "192.168.3.0/24", "192.168.4.0/25", "192.168.101.0/26", "192.168.102.1/27"})
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		ipaddr.Supernet(subs)
	}
}

func BenchmarkAggregateIPv4(b *testing.B) {
	subs, err := toPrefixes([]string{"192.168.0.0/24", "192.168.1.0/24", "192.168.2.0/24", "192.168.3.0/24", "192.168.4.0/25", "192.168.101.0/26", "192.168.102.1/27"})
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		ipaddr.Aggregate(subs)
	}
}

func BenchmarkSummarizeIPv4(b *testing.B) {
	first, last := net.IPv4(172, 16, 1, 1), net.IPv4(172, 16, 255, 255)
	for i := 0; i < b.N; i++ {
		ipaddr.Summarize(first, last)
	}
}
