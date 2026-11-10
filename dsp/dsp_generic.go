// Copyright 2020 Justine Alexandra Roberts Tunney
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Portable implementations of the DSP primitives. The upstream module only
// ships dsp_amd64.s, so on other architectures (e.g. arm64) the declarations
// in dsp.go would otherwise have no function bodies.

//go:build !amd64

package dsp

import (
	"math/bits"
	"unsafe"
)

const ulawBias = 0x84

// L16MixSat160 mixes src into dst using saturating 16-bit addition.
func L16MixSat160(dst, src *int16) {
	d := (*[160]int16)(unsafe.Pointer(dst))
	s := (*[160]int16)(unsafe.Pointer(src))
	for i := 0; i < 160; i++ {
		v := int32(d[i]) + int32(s[i])
		switch {
		case v > 32767:
			v = 32767
		case v < -32768:
			v = -32768
		}
		d[i] = int16(v)
	}
}

// LinearToUlaw compresses a PCM audio sample into a G.711 mu-Law sample.
func LinearToUlaw(linear int64) int64 {
	var mask int64
	if linear < 0 {
		linear = ulawBias - linear
		mask = 0x7F
	} else {
		linear += ulawBias
		mask = 0xFF
	}
	// Same as BSR on (linear | 0xFF), minus 7.
	exponent := int64(bits.Len64(uint64(linear|0xFF))) - 1 - 7
	if exponent >= 8 {
		return mask ^ 0x7F
	}
	mantissa := (linear >> uint(exponent+3)) & 0x0F
	return ((exponent << 4) | mantissa) ^ mask
}

// UlawToLinear turns a G.711 mu-Law sample back into an audio sample.
func UlawToLinear(ulaw int64) int64 {
	u := ^ulaw
	mantissa := u & 0x0F
	exponent := (u & 0x70) >> 4
	linear := ((mantissa << 3) + ulawBias) << uint(exponent)
	if u&0x80 != 0 {
		return ulawBias - linear
	}
	return linear - ulawBias
}
