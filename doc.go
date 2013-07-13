// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

// Package ipaddr provides basic functions for the manipulation of IP
// address prefixes and subsequent addresses as described in RFC 4632
// and RFC 4291.
//
//
// Examples:
//
//	i := big.NewInt(530000)
//	p, err := ipaddr.NewPrefix(net.ParseIP("10.128.0.0"), ipaddr.IPv4PrefixLen-i.BitLen())
//	if err != nil {
//		// error handling
//	}
//	for _, sub := range p.Subnets(3)
//		subs := sub.Subnets(2) {
//		for _, sub := range subs {
//			if sub.Bits(sub.Len()-1, 1)&0x1 != 0 {
//				// do something
//			}
//		}
//		sum := ipaddr.SummaryPrefix(subs[i:j])
//	}
package ipaddr
