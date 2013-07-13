// Copyright 2013 Mikio Hara. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.

package ipaddr

import (
	"errors"
	"io"
	"net"
)

var (
	errUnknownDecodingType = errors.New("unknown decoding type")
	errUnknownEncodingType = errors.New("unknown encoding type")
	errShortWrite          = errors.New("short write")
)

// A Decoder represents an IP address prefix decoder.
type Decoder struct {
	r  io.Reader
	fn func(*[]Prefix) error
}

// NewDecoder returns a new decoder that reads from r. Known address
// family identifiers are "ipv4" and "ipv6", known decoding type is
// "nlri" (Network Layer Reachability Information).
func NewDecoder(afi, typ string, r io.Reader) (*Decoder, error) {
	dec := &Decoder{r: r}
	switch {
	case afi == "ipv4" && typ == "nlri":
		dec.fn = dec.decodeIPv4NLRI
	case afi == "ipv6" && typ == "nlri":
		dec.fn = dec.decodeIPv6NLRI
	default:
		return nil, errUnknownDecodingType
	}
	return dec, nil
}

// Decode reads and parses the specified decoding type of prefixes
// from the stream.
func (dec *Decoder) Decode(prefixes *[]Prefix) error {
	return dec.fn(prefixes)
}

func (dec *Decoder) decodeIPv4NLRI(prefixes *[]Prefix) error {
	for {
		var b [1 + net.IPv4len]byte
		n, err := dec.r.Read(b[:])
		if n < 1+net.IPv4len {
			return nil
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		*prefixes = append(*prefixes, nlriToIPv4(b[:]))
	}
}

func (dec *Decoder) decodeIPv6NLRI(prefixes *[]Prefix) error {
	for {
		var b [1 + net.IPv6len]byte
		n, err := dec.r.Read(b[:])
		if n < 1+net.IPv6len {
			return nil
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		*prefixes = append(*prefixes, nlriToIPv6(b[:]))
	}
}

// An Encoder represents an IP address prefix encoder.
type Encoder struct {
	w  io.Writer
	fn func([]Prefix) error
}

// NewEncoder returns a new encoder that writes to w. Known address
// family identifiers are "ipv4" and "ipv6", known encoding type is
// "nlri" (Network Layer Reachability Information).
func NewEncoder(afi, typ string, w io.Writer) (*Encoder, error) {
	enc := &Encoder{w: w}
	switch {
	case afi == "ipv4" && typ == "nlri":
		enc.fn = enc.encodeIPv4NLRI
	case afi == "ipv6" && typ == "nlri":
		enc.fn = enc.encodeIPv6NLRI
	default:
		return nil, errUnknownEncodingType
	}
	return enc, nil
}

// Encode writes the specified encoding type of prefixes to the
// stream.
func (enc *Encoder) Encode(prefixes []Prefix) error {
	return enc.fn(prefixes)
}

func (enc *Encoder) encodeIPv4NLRI(prefixes []Prefix) error {
	var b [1 + net.IPv4len]byte
	for _, p := range prefixes {
		p, ok := p.(*IPv4)
		if !ok {
			continue
		}
		if n, err := enc.w.Write(p.addr.encodeNLRI(b[:], p.nbits)); err != nil {
			return err
		} else if n < 1+net.IPv4len {
			return errShortWrite
		}
	}
	return nil
}

func (enc *Encoder) encodeIPv6NLRI(prefixes []Prefix) error {
	var b [1 + net.IPv6len]byte
	for _, p := range prefixes {
		p, ok := p.(*IPv6)
		if !ok {
			continue
		}
		if n, err := enc.w.Write(p.addr.encodeNLRI(b[:], p.nbits)); err != nil {
			return err
		} else if n < 1+net.IPv6len {
			return errShortWrite
		}
	}
	return nil
}
