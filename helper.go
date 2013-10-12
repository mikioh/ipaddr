// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr

import (
	"encoding/binary"
	"math"
	"net"
)

func npop64(bs uint64) byte {
	bs = bs&0x5555555555555555 + bs>>1&0x5555555555555555
	bs = bs&0x3333333333333333 + bs>>2&0x3333333333333333
	bs = bs&0x0f0f0f0f0f0f0f0f + bs>>4&0x0f0f0f0f0f0f0f0f
	bs = bs&0x00ff00ff00ff00ff + bs>>8&0x00ff00ff00ff00ff
	bs = bs&0x0000ffff0000ffff + bs>>16&0x0000ffff0000ffff
	return byte(bs&0x00000000ffffffff + bs>>32&0x00000000ffffffff)
}

func npop32(bs uint32) byte {
	bs = bs&0x55555555 + bs>>1&0x55555555
	bs = bs&0x33333333 + bs>>2&0x33333333
	bs = bs&0x0f0f0f0f + bs>>4&0x0f0f0f0f
	bs = bs&0x00ff00ff + bs>>8&0x00ff00ff
	return byte(bs&0x0000ffff + bs>>16&0x0000ffff)
}

func nlz64(bs uint64) byte {
	bs |= bs >> 1
	bs |= bs >> 2
	bs |= bs >> 4
	bs |= bs >> 8
	bs |= bs >> 16
	bs |= bs >> 32
	return npop64(^bs)
}

func nlz32(bs uint32) byte {
	bs |= bs >> 1
	bs |= bs >> 2
	bs |= bs >> 4
	bs |= bs >> 8
	bs |= bs >> 16
	return npop32(^bs)
}

func mask64(nbits byte) uint64 {
	return -uint64(1 << uint(64-nbits))
}

func mask32(nbits byte) uint32 {
	return -uint32(1 << uint(32-nbits))
}

type ipv4Int uint32

func (i ipv4Int) IP() net.IP {
	var b [net.IPv4len]byte
	binary.BigEndian.PutUint32(b[:], uint32(i))
	return net.IP(b[:])
}

func (i ipv4Int) encodeNLRI(b []byte, nbits byte) int {
	l := int(nbits) / 8
	if nbits%8 > 0 {
		l++
	}
	b[0] = nbits
	l++
	b = b[1:]
	for n := 0; n < l; n++ {
		b[n] = byte(i >> uint(32-8*(n+1)))
	}
	return l
}

func (p *IPv4) decodeNLRI(b []byte) error {
	p.nbits = b[0]
	l := len(b) - 1
	b = b[1:]
	for n := 0; n < l; n++ {
		p.addr |= ipv4Int(b[n]) << uint(32-8*(n+1))
	}
	return nil
}

func ipToIPv4Int(ip []byte) ipv4Int {
	return ipv4Int(binary.BigEndian.Uint32(ip))
}

type ipv6Int [2]uint64

func (i *ipv6Int) lshift(nbits byte) {
	i[0] = i[0]<<uint(nbits) | i[1]>>uint(64-nbits) | i[1]<<uint(nbits-64)
	i[1] = i[1] << uint(nbits)
}

func (i *ipv6Int) rshift(nbits byte) {
	i[1] = i[1]>>uint(nbits) | i[0]<<uint(64-nbits) | i[0]>>uint(nbits-64)
	i[0] = i[0] >> uint(nbits)
}

func (i *ipv6Int) incr() {
	if i[1] == math.MaxUint64 {
		i[0]++
		i[1] = 0
	} else {
		i[1]++
	}
}

func (i *ipv6Int) setHostmask(nbits byte) {
	if nbits > 64 {
		i[0], i[1] = ^mask64(64), ^mask64(nbits-64)
	} else {
		i[0], i[1] = ^mask64(nbits), mask64(64)
	}
}

func (i *ipv6Int) setNetmask(nbits byte) {
	if nbits > 64 {
		i[0], i[1] = mask64(64), mask64(nbits-64)
	} else {
		i[0], i[1] = mask64(nbits), 0
	}
}

func (i *ipv6Int) IP() net.IP {
	var b [net.IPv6len]byte
	binary.BigEndian.PutUint64(b[:8], i[0])
	binary.BigEndian.PutUint64(b[8:16], i[1])
	return net.IP(b[:])
}

func (i *ipv6Int) encodeNLRI(b []byte, nbits byte) int {
	l := int(nbits) / 8
	if nbits%8 > 0 {
		l++
	}
	b[0] = nbits
	l++
	b = b[1:]
	for n := 0; n < l; n++ {
		switch {
		case n < 8:
			b[n] = byte(i[0] >> uint(64-(8*(n+1))))
		case 8 <= n && n < 16:
			b[n] = byte(i[1] >> uint(128-(8*(n+1))))
		}
	}
	return l
}

func (p *IPv6) decodeNLRI(b []byte) error {
	p.nbits = b[0]
	l := len(b) - 1
	b = b[1:]
	for n := 0; n < l; n++ {
		switch {
		case n < 8:
			p.addr[0] |= uint64(b[n]) << uint(64-8*(n+1))
		case 8 <= n && n < 16:
			p.addr[1] |= uint64(b[n]) << uint(128-8*(n+1))
		}
	}
	return nil
}

func ipToIPv6Int(ip []byte) ipv6Int {
	return ipv6Int{binary.BigEndian.Uint64(ip[:8]), binary.BigEndian.Uint64(ip[8:16])}
}
