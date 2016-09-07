// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr

import (
	"bytes"
	"sort"
)

type byAddrFamily []Prefix

func (ps byAddrFamily) ipv4Only() []Prefix {
	nps := make([]Prefix, 0, len(ps))
	for _, p := range ps {
		if p.IP.To4() != nil {
			nps = append(nps, p)
		}
	}
	return nps
}

func (ps byAddrFamily) ipv6Only() []Prefix {
	nps := make([]Prefix, 0, len(ps))
	for _, p := range ps {
		if p.IP.To16() != nil && p.IP.To4() == nil {
			nps = append(nps, p)
		}
	}
	return nps
}

type byAscending []Prefix

func (ps byAscending) Len() int           { return len(ps) }
func (ps byAscending) Less(i, j int) bool { return compareAscending(&ps[i], &ps[j]) < 0 }
func (ps byAscending) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }

func compareAscending(a, b *Prefix) int {
	if n := bytes.Compare(a.IP, b.IP); n != 0 {
		return n
	}
	if n := bytes.Compare(a.Mask, b.Mask); n != 0 {
		return n
	}
	return 0
}

type byDescending []Prefix

func (ps byDescending) Len() int           { return len(ps) }
func (ps byDescending) Less(i, j int) bool { return compareDescending(&ps[i], &ps[j]) >= 0 }
func (ps byDescending) Swap(i, j int)      { ps[i], ps[j] = ps[j], ps[i] }

func compareDescending(a, b *Prefix) int {
	if n := bytes.Compare(a.Mask, b.Mask); n != 0 {
		return n
	}
	if n := bytes.Compare(a.IP, b.IP); n != 0 {
		return n
	}
	return 0
}

func sortAndDedup(ps []Prefix, dir int, strict bool) []Prefix {
	if len(ps) == 0 {
		return nil
	}
	if strict {
		if ps[0].IP.To4() != nil {
			ps = byAddrFamily(ps).ipv4Only()
			if dir == 'a' {
				sort.Sort(byAscending(ps))
			} else {
				sort.Sort(byDescending(ps))
			}
		}
		if ps[0].IP.To16() != nil && ps[0].IP.To4() == nil {
			ps = byAddrFamily(ps).ipv6Only()
			if dir == 'a' {
				sort.Sort(byAscending(ps))
			} else {
				sort.Sort(byDescending(ps))
			}
		}
	} else {
		pps := make([]Prefix, 0, len(ps))
		for _, p := range ps {
			pps = append(pps, p)
		}
		if dir == 'a' {
			sort.Sort(byAscending(pps))
		} else {
			sort.Sort(byDescending(pps))
		}
		ps = pps
	}
	nps := ps[:0]
	var prev *Prefix
	for i := range ps {
		if prev == nil {
			nps = append(nps, ps[i])
		} else if !prev.Equal(&ps[i]) {
			nps = append(nps, ps[i])
		}
		prev = &ps[i]
	}
	return nps
}
