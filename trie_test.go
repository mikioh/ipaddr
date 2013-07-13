// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr_test

import (
	"github.com/mikioh/ipaddr"
	"net"
	"sort"
	"testing"
)

var prefixTrieTests = [][]struct {
	addr      net.IP
	prefixLen int
}{
	{
		{net.ParseIP("0.0.0.0"), 4},
		{net.ParseIP("16.0.0.0"), 4},
		{net.ParseIP("40.0.0.0"), 5},
		{net.ParseIP("64.0.0.0"), 3},
		{net.ParseIP("96.0.0.0"), 4},
		{net.ParseIP("112.0.0.0"), 4},
		{net.ParseIP("128.0.0.0"), 3},
		{net.ParseIP("160.0.0.0"), 6},
		{net.ParseIP("164.0.0.0"), 6},
		{net.ParseIP("168.0.0.0"), 5},
		{net.ParseIP("176.0.0.0"), 5},
		{net.ParseIP("184.0.0.0"), 5},
		{net.ParseIP("192.0.0.0"), 3},
		{net.ParseIP("232.0.0.0"), 8},
		{net.ParseIP("233.0.0.0"), 8},
	},
}

func TestPrefixTrie(t *testing.T) {
	for _, tt := range prefixTrieTests {
		var ps []ipaddr.Prefix
		for _, st := range tt {
			p, err := ipaddr.NewPrefix(st.addr, st.prefixLen)
			if err != nil {
				t.Fatalf("ipaddr.NewPrefix failed: %v", err)
			}
			ps = append(ps, p)
		}
		sort.Sort(byAddress(ps))
		ni := 1
		ns := prefixTrie(make([]ipv4Node, 0), &ni, ps, 0, 0)
		traverse(t, ns, ps, ns[0])
	}
}

func traverse(t *testing.T, ns []ipv4Node, ps []ipaddr.Prefix, n ipv4Node) {
	if n.bfac() == 0 {
		t.Logf("leaf: %v for %v", n, ps[n.index()])
		return
	}
	for i := 0; i < 1<<uint(n.bfac()); i++ {
		traverse(t, ns, ps, ns[n.index()+i])
	}
}

type byAddress []ipaddr.Prefix

func (ps byAddress) Len() int {
	return len(ps)
}

func (ps byAddress) Less(i, j int) bool {
	if ipaddr.ComparePrefix(ps[i], ps[j]) < 0 {
		return true
	}
	return false
}

func (ps byAddress) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

type byPrefixLen []ipaddr.Prefix

func (ps byPrefixLen) Len() int {
	return len(ps)
}

func (ps byPrefixLen) Less(i, j int) bool {
	if ps[i].Len() < ps[j].Len() {
		return true
	} else if ps[i].Len() > ps[j].Len() {
		return false
	}
	if ipaddr.ComparePrefix(ps[i], ps[j]) < 0 {
		return true
	}
	return false
}

func (ps byPrefixLen) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}
