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

func (i ipv4Int) toIP() net.IP {
	var b [net.IPv4len]byte
	binary.BigEndian.PutUint32(b[:], uint32(i))
	return net.IP(b[:])
}

func (i ipv4Int) encodeNLRI(b []byte, nbits byte) []byte {
	b[0] = nbits
	binary.BigEndian.PutUint32(b[1:5], uint32(i))
	return b[:5]
}

func ipToIPv4Int(ip []byte) ipv4Int {
	return ipv4Int(binary.BigEndian.Uint32(ip))
}

func nlriToIPv4(b []byte) *IPv4 {
	return &IPv4{nbits: b[0], addr: ipv4Int(binary.BigEndian.Uint32(b[1:5]))}
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

func (i *ipv6Int) toIP() net.IP {
	var b [net.IPv6len]byte
	binary.BigEndian.PutUint64(b[:8], i[0])
	binary.BigEndian.PutUint64(b[8:16], i[1])
	return net.IP(b[:])
}

func (i *ipv6Int) encodeNLRI(b []byte, nbits byte) []byte {
	b[0] = nbits
	binary.BigEndian.PutUint64(b[1:9], i[0])
	binary.BigEndian.PutUint64(b[9:17], i[1])
	return b[:17]
}

func ipToIPv6Int(ip []byte) ipv6Int {
	return ipv6Int{binary.BigEndian.Uint64(ip[:8]), binary.BigEndian.Uint64(ip[8:16])}
}

func nlriToIPv6(b []byte) *IPv6 {
	return &IPv6{nbits: b[0], addr: ipv6Int{binary.BigEndian.Uint64(b[1:9]), binary.BigEndian.Uint64(b[9:17])}}
}
