// Copyright 2015 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr_test

import (
	"fmt"
	"log"
	"net"

	"github.com/mikioh/ipaddr"
)

func ExamplePrefix_subnettingAndSupernetting() {
	_, ipn, err := net.ParseCIDR("172.16.0.0/16")
	if err != nil {
		log.Fatal(err)
	}
	nbits, _ := ipn.Mask.Size()
	p, err := ipaddr.NewPrefix(ipn.IP, nbits)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(p.Addr(), p.LastAddr(), p.Len(), p.Netmask(), p.Hostmask())

	subs := p.Subnets(3)
	for _, sub := range subs {
		fmt.Println(sub)
	}

	fmt.Println(ipaddr.Supernet(subs[4:6]))

	// Output:
	// 172.16.0.0 172.16.255.255 16 ffff0000 0000ffff
	// 172.16.0.0/19
	// 172.16.32.0/19
	// 172.16.64.0/19
	// 172.16.96.0/19
	// 172.16.128.0/19
	// 172.16.160.0/19
	// 172.16.192.0/19
	// 172.16.224.0/19
	// 172.16.128.0/18
}

func ExamplePrefix_subnettingAndAggregation() {
	_, ipn, err := net.ParseCIDR("192.168.0.0/24")
	if err != nil {
		log.Fatal(err)
	}
	nbits, _ := ipn.Mask.Size()
	p, err := ipaddr.NewPrefix(ipn.IP, nbits)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(p.Addr(), p.LastAddr(), p.Len(), p.Netmask(), p.Hostmask())

	subs := p.Subnets(2)
	for _, sub := range subs {
		fmt.Println(sub)
	}
	px, err := ipaddr.NewPrefix(net.IPv4(192, 168, 100, 0), 24)
	if err != nil {
		log.Fatal(err)
	}
	subs = append(subs, px)

	fmt.Println(ipaddr.Aggregate(subs[2:]))

	// Output:
	// 192.168.0.0 192.168.0.255 24 ffffff00 000000ff
	// 192.168.0.0/26
	// 192.168.0.64/26
	// 192.168.0.128/26
	// 192.168.0.192/26
	// [192.168.0.128/25 192.168.100.0/24]
}
