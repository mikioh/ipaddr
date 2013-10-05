// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

// +build go1.2

package ipaddr

import "encoding"

var (
	_ encoding.TextMarshaler   = &IPv4{}
	_ encoding.TextUnmarshaler = &IPv4{}
	_ encoding.TextMarshaler   = &IPv6{}
	_ encoding.TextUnmarshaler = &IPv6{}
)
