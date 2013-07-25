// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr_test

import (
	"bytes"
	"net"
	"testing"

	"github.com/mikioh/ipaddr"
)

var nlriCodecTests = []struct {
	afi, typ string
	ss       []string
}{
	{"ipv4", "nlri", []string{
		"192.168.1.0/23",
		"192.168.2.0/24",
		"192.168.3.0/24",
		"192.168.4.255/32",
	}},

	{"ipv6", "nlri", []string{
		"2001:db8:1::/47",
		"2001:db8:2::/48",
		"2001:db8:3::/48",
		"2001:db8:4::1/128",
	}},
}

func TestNLRICodec(t *testing.T) {
	for i, tt := range nlriCodecTests {
		var (
			src []byte
			ns  []*net.IPNet
		)
		for _, s := range tt.ss {
			b, n, err := encodeNLRI(s)
			if err != nil {
				continue
			}
			src, ns = append(src, b...), append(ns, n)
		}
		dec, err := ipaddr.NewDecoder(tt.afi, tt.typ, bytes.NewReader(src))
		if err != nil {
			t.Fatalf("ipaddr.NewDecoder failed: %v", err)
		}
		var ps []ipaddr.Prefix
		if err := dec.Decode(&ps); err != nil {
			t.Fatalf("ipaddr.Decoder.Decode failed: %v", err)
		}
		for i, p := range ps {
			prefixLen, _ := ns[i].Mask.Size()
			if !p.Addr().Equal(ns[i].IP) || p.Len() != prefixLen {
				t.Fatalf("#%v: got %v; expected %v", i, p, ns[i])
			}
		}
		var dst bytes.Buffer
		enc, err := ipaddr.NewEncoder(tt.afi, tt.typ, &dst)
		if err != nil {
			t.Fatalf("ipaddr.NewEncoder failed: %v", err)
		}
		if err := enc.Encode(ps); err != nil {
			t.Fatalf("ipaddr.Encoder.Encode failed: %v", err)
		}
		if bytes.Compare(dst.Bytes(), src) != 0 {
			t.Fatalf("#%v: encoded output looks wrong", i)
		}
	}
}

func encodeNLRI(ns string) ([]byte, *net.IPNet, error) {
	_, n, err := net.ParseCIDR(ns)
	if err != nil {
		return nil, nil, err
	}
	prefixLen, size := n.Mask.Size()
	switch size {
	case ipaddr.IPv4PrefixLen:
		return append([]byte{byte(prefixLen)}, n.IP.To4()...), n, nil
	case ipaddr.IPv6PrefixLen:
		return append([]byte{byte(prefixLen)}, n.IP.To16()...), n, nil
	}
	return nil, n, nil
}
