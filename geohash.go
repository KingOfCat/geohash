// Package geohash provides encoding and decoding of string and integer
// geohashes.
package geohash

import (
	"fmt"
	"math"
)

// Encode the point (lat, lng) as a string geohash with the standard 12
// characters of precision.
func Encode(lat, lng float64) string {
	return EncodeWithPrecision(lat, lng, 12)
}

// EncodeWithPrecision encodes the point (lat, lng) as a string geohash with
// the specified number of characters of precision (max 12).
func EncodeWithPrecision(lat, lng float64, chars uint) string {
	bits := 5 * chars
	inthash := EncodeIntWithPrecision(lat, lng, bits)
	enc := base32encoding.Encode(inthash)
	return enc[12-chars:]
}

// encodeIntGeneric provides a Go implementation of integer geohash. The
// definition of the public EncodeInt function is in architecture specific
// files.
func encodeIntGeneric(lat, lng float64) uint64 {
	fmt.Println("generic")
	latInt := encodeRange(lat, 90)
	lngInt := encodeRange(lng, 180)
	return interleave(latInt, lngInt)
}

// EncodeIntWithPrecision encodes the point (lat, lng) to an integer with the
// specified number of bits.
func EncodeIntWithPrecision(lat, lng float64, bits uint) uint64 {
	hash := EncodeInt(lat, lng)
	return hash >> (64 - bits)
}

// Box represents a rectangle in latitude/longitude space.
type Box struct {
	MinLat float64
	MaxLat float64
	MinLng float64
	MaxLng float64
}

// Center returns the center of the box.
func (b Box) Center() (lat, lng float64) {
	lat = (b.MinLat + b.MaxLat) / 2.0
	lng = (b.MinLng + b.MaxLng) / 2.0
	return
}

// Contains decides whether (lat, lng) is contained in the box. The
// containment test is inclusive of the edges and corners.
func (b Box) Contains(lat, lng float64) bool {
	return (b.MinLat <= lat && lat <= b.MaxLat &&
		b.MinLng <= lng && lng <= b.MaxLng)
}

// minDecimalPlaces returns the minimum number of decimal places such that
// there must exist an number with that many places within any range of width
// r. This is intended for returning minimal precision coordinates inside a
// box.
func maxDecimalPower(r float64) float64 {
	m := int(math.Floor(math.Log10(r)))
	return math.Pow10(m)
}

// Round returns a point inside the box, making an effort to round to minimal
// precision.
func (b Box) Round() (lat, lng float64) {
	x := maxDecimalPower(b.MaxLat - b.MinLat)
	lat = math.Ceil(b.MinLat/x) * x
	x = maxDecimalPower(b.MaxLng - b.MinLng)
	lng = math.Ceil(b.MinLng/x) * x
	return
}

// errorWithPrecision returns the error range in latitude and longitude for in
// integer geohash with bits of precision.
func errorWithPrecision(bits uint) (latErr, lngErr float64) {
	latBits := bits / 2
	lngBits := bits - latBits
	latErr = 180.0 / math.Exp2(float64(latBits))
	lngErr = 360.0 / math.Exp2(float64(lngBits))
	return
}

// BoundingBox returns the region encoded by the given string geohash.
func BoundingBox(hash string) Box {
	bits := uint(5 * len(hash))
	inthash := base32encoding.Decode(hash)
	return BoundingBoxIntWithPrecision(inthash, bits)
}

// BoundingBoxIntWithPrecision returns the region encoded by the integer
// geohash with the specified precision.
func BoundingBoxIntWithPrecision(hash uint64, bits uint) Box {
	fullHash := hash << (64 - bits)
	latInt, lngInt := deinterleave(fullHash)
	lat := decodeRange(latInt, 90)
	lng := decodeRange(lngInt, 180)
	latErr, lngErr := errorWithPrecision(bits)
	return Box{
		MinLat: lat,
		MaxLat: lat + latErr,
		MinLng: lng,
		MaxLng: lng + lngErr,
	}
}

// BoundingBoxInt returns the region encoded by the given 64-bit integer
// geohash.
func BoundingBoxInt(hash uint64) Box {
	return BoundingBoxIntWithPrecision(hash, 64)
}

// Decode the string geohash to a (lat, lng) point.
func Decode(hash string) (lat, lng float64) {
	box := BoundingBox(hash)
	return box.Round()
}

// DecodeIntWithPrecision decodes the provided integer geohash with bits of
// precision to a (lat, lng) point.
func DecodeIntWithPrecision(hash uint64, bits uint) (lat, lng float64) {
	box := BoundingBoxIntWithPrecision(hash, bits)
	return box.Round()
}

// DecodeInt decodes the provided 64-bit integer geohash to a (lat, lng) point.
func DecodeInt(hash uint64) (lat, lng float64) {
	return DecodeIntWithPrecision(hash, 64)
}

// Neighbors returns a slice of geohash strings that correspond to the provided
// geohash's neighbors.
func Neighbors(hash string) []string {
	box := BoundingBox(hash)
	lat, lng := box.Center()
	latDelta := box.MaxLat - box.MinLat
	lngDelta := box.MaxLng - box.MinLng
	precision := uint(len(hash))
	return []string{
		// N
		EncodeWithPrecision(lat+latDelta, lng, precision),
		// NE,
		EncodeWithPrecision(lat+latDelta, lng+lngDelta, precision),
		// E,
		EncodeWithPrecision(lat, lng+lngDelta, precision),
		// SE,
		EncodeWithPrecision(lat-latDelta, lng+lngDelta, precision),
		// S,
		EncodeWithPrecision(lat-latDelta, lng, precision),
		// SW,
		EncodeWithPrecision(lat-latDelta, lng-lngDelta, precision),
		// W,
		EncodeWithPrecision(lat, lng-lngDelta, precision),
		// NW
		EncodeWithPrecision(lat+latDelta, lng-lngDelta, precision),
	}
}

// NeighborsInt returns a slice of uint64s that correspond to the provided hash's
// neighbors at 64-bit precision.
func NeighborsInt(hash uint64) []uint64 {
	return NeighborsIntWithPrecision(hash, 64)
}

// NeighborsIntWithPrecision returns a slice of uint64s that correspond to the
// provided hash's neighbors at the given precision.
func NeighborsIntWithPrecision(hash uint64, bits uint) []uint64 {
	box := BoundingBoxIntWithPrecision(hash, bits)
	lat, lng := box.Center()
	latDelta := box.MaxLat - box.MinLat
	lngDelta := box.MaxLng - box.MinLng
	return []uint64{
		// N
		EncodeIntWithPrecision(lat+latDelta, lng, bits),
		// NE,
		EncodeIntWithPrecision(lat+latDelta, lng+lngDelta, bits),
		// E,
		EncodeIntWithPrecision(lat, lng+lngDelta, bits),
		// SE,
		EncodeIntWithPrecision(lat-latDelta, lng+lngDelta, bits),
		// S,
		EncodeIntWithPrecision(lat-latDelta, lng, bits),
		// SW,
		EncodeIntWithPrecision(lat-latDelta, lng-lngDelta, bits),
		// W,
		EncodeIntWithPrecision(lat, lng-lngDelta, bits),
		// NW
		EncodeIntWithPrecision(lat+latDelta, lng-lngDelta, bits),
	}
}

// precalculated for performance
var exp232 = math.Exp2(32)

// Encode the position of x within the range -r to +r as a 32-bit integer.
func encodeRange(x, r float64) uint32 {
	p := (x + r) / (2 * r)
	return uint32(p * exp232)
}

// Decode the 32-bit range encoding X back to a value in the range -r to +r.
func decodeRange(X uint32, r float64) float64 {
	p := float64(X) / exp232
	x := 2*r*p - r
	return x
}

// Spread out the 32 bits of x into 64 bits, where the bits of x occupy even
// bit positions.
func spread(x uint32) uint64 {
	X := uint64(x)
	X = (X | (X << 16)) & 0x0000ffff0000ffff
	X = (X | (X << 8)) & 0x00ff00ff00ff00ff
	X = (X | (X << 4)) & 0x0f0f0f0f0f0f0f0f
	X = (X | (X << 2)) & 0x3333333333333333
	X = (X | (X << 1)) & 0x5555555555555555
	return X
}

// Interleave the bits of x and y. In the result, x and y occupy even and odd
// bitlevels, respectively.
func interleave(x, y uint32) uint64 {
	return spread(x) | (spread(y) << 1)
}

// Squash the even bitlevels of X into a 32-bit word. Odd bitlevels of X are
// ignored, and may take any value.
func squash(X uint64) uint32 {
	X &= 0x5555555555555555
	X = (X | (X >> 1)) & 0x3333333333333333
	X = (X | (X >> 2)) & 0x0f0f0f0f0f0f0f0f
	X = (X | (X >> 4)) & 0x00ff00ff00ff00ff
	X = (X | (X >> 8)) & 0x0000ffff0000ffff
	X = (X | (X >> 16)) & 0x00000000ffffffff
	return uint32(X)
}

// Deinterleave the bits of X into 32-bit words containing the even and odd
// bitlevels of X, respectively.
func deinterleave(X uint64) (uint32, uint32) {
	return squash(X), squash(X >> 1)
}
