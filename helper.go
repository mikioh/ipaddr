// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr

import "math/bits"

func mask32(nbits int) uint32 {
	return -uint32(1 << uint(32-nbits))
}

func mask64(nbits int) uint64 {
	return -uint64(1 << uint(64-nbits))
}

func leadingZeros32(bs uint32) int {
	return int(bits.LeadingZeros32(bs))
}

func leadingZeros64(bs uint64) int {
	return int(bits.LeadingZeros64(bs))
}
