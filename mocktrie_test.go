// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr_test

import (
	"fmt"

	"github.com/mikioh/ipaddr"
)

func npop32(bs uint32) int {
	bs = bs&0x55555555 + bs>>1&0x55555555
	bs = bs&0x33333333 + bs>>2&0x33333333
	bs = bs&0x0f0f0f0f + bs>>4&0x0f0f0f0f
	bs = bs&0x00ff00ff + bs>>8&0x00ff00ff
	return int(bs&0x0000ffff + bs>>16&0x0000ffff)
}

func nlz32(bs uint32) int {
	bs |= bs >> 1
	bs |= bs >> 2
	bs |= bs >> 4
	bs |= bs >> 8
	bs |= bs >> 16
	return npop32(^bs)
}

// An ipv4Node represents a 32-bit trie node consists of 5-bit
// branching factor, 7-bit skipping length and 20-bit index to the
// stored prefix.
type ipv4Node uint32

func (n ipv4Node) String() string {
	return fmt.Sprintf("[bf: %x, off: %d, i: %d]", n.bfac(), n.nskip(), n.index())
}

func (n ipv4Node) bfac() int {
	return int(n >> 27 & 0x1f)
}

func (n ipv4Node) nskip() int {
	return int(n >> 20 & 0x7f)
}

func (n ipv4Node) index() int {
	return int(n & 0x000fffff)
}

func (n ipv4Node) bits(pos, nbits int) uint32 {
	return uint32(n) << uint(pos) >> uint(ipaddr.IPv4PrefixLen-nbits)
}

func newIPv4Node(f, n, i int) ipv4Node {
	return ipv4Node(f)<<27 | ipv4Node(n)<<20 | ipv4Node(i)&0x000fffff
}

func nskip(ps []ipaddr.Prefix, pos int) int {
	first := ps[0].Bits(pos, ipaddr.IPv4PrefixLen-pos)
	last := ps[len(ps)-1].Bits(pos, ipaddr.IPv4PrefixLen-pos)
	return int(nlz32(first^last)) - pos
}

func bfac(ps []ipaddr.Prefix, pos int) int {
	if len(ps) == 2 {
		return 1
	}
	bf := 2 // start the search with 4 (2^bf) branches
	for ; pos+bf <= ipaddr.IPv4PrefixLen; bf++ {
		nsubs := 0
		for pat := uint32(0); pat < 1<<uint(bf); pat++ {
			for _, p := range ps {
				if pat == p.Bits(pos, bf) {
					nsubs++
					break
				}
			}
		}
		if nsubs < 1<<uint(bf) {
			break
		}
	}
	return bf - 1
}

func index(ns []ipv4Node, t ipv4Node) int {
	pos, bf, i := ns[0].nskip(), ns[0].bfac(), ns[0].index()
	for bf != 0 {
		n := ns[i+int(t.bits(pos, bf))]
		pos += bf + n.nskip()
		bf, i = n.bfac(), n.index()
	}
	return i
}

func prefixTrie(ns []ipv4Node, nindex *int, ps []ipaddr.Prefix, pindex, pos int) []ipv4Node {
	if len(ps) == 0 {
		return nil
	} else if len(ps) == 1 {
		ns = append(ns, newIPv4Node(0, 0, pindex))
		return ns
	}
	off := nskip(ps, pos)
	bf := bfac(ps, pos+off)
	ns = append(ns, newIPv4Node(bf, off, *nindex))
	*nindex += 1 << uint(bf)
	for pat := uint32(0); pat < 1<<uint(bf); pat++ {
		nsubs := 0
		for _, p := range ps {
			if pat != p.Bits(pos+off, bf) {
				break
			}
			nsubs++
		}
		ns = prefixTrie(ns, nindex, ps[:nsubs], pindex, pos+off+bf)
		ps = ps[nsubs:]
		pindex += nsubs
	}
	return ns
}
