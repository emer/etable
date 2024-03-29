// Code generated by "core generate"; DO NOT EDIT.

package etensor

import (
	"cogentcore.org/core/enums"
)

var _TypeValues = []Type{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}

// TypeN is the highest valid value for type Type, plus one.
const TypeN Type = 15

var _TypeValueMap = map[string]Type{`NULL`: 0, `BOOL`: 1, `UINT8`: 2, `INT8`: 3, `UINT16`: 4, `INT16`: 5, `UINT32`: 6, `INT32`: 7, `UINT64`: 8, `INT64`: 9, `FLOAT16`: 10, `FLOAT32`: 11, `FLOAT64`: 12, `STRING`: 13, `INT`: 14}

var _TypeDescMap = map[Type]string{0: `Null type having no physical storage`, 1: `Bool is a 1 bit, LSB bit-packed ordering`, 2: `UINT8 is an Unsigned 8-bit little-endian integer`, 3: `INT8 is a Signed 8-bit little-endian integer`, 4: `UINT16 is an Unsigned 16-bit little-endian integer`, 5: `INT16 is a Signed 16-bit little-endian integer`, 6: `UINT32 is an Unsigned 32-bit little-endian integer`, 7: `INT32 is a Signed 32-bit little-endian integer`, 8: `UINT64 is an Unsigned 64-bit little-endian integer`, 9: `INT64 is a Signed 64-bit little-endian integer`, 10: `FLOAT16 is a 2-byte floating point value`, 11: `FLOAT32 is a 4-byte floating point value`, 12: `FLOAT64 is an 8-byte floating point value`, 13: `STRING is a UTF8 variable-length string`, 14: `INT is a Signed 64-bit little-endian integer -- should only use on 64bit machines!`}

var _TypeMap = map[Type]string{0: `NULL`, 1: `BOOL`, 2: `UINT8`, 3: `INT8`, 4: `UINT16`, 5: `INT16`, 6: `UINT32`, 7: `INT32`, 8: `UINT64`, 9: `INT64`, 10: `FLOAT16`, 11: `FLOAT32`, 12: `FLOAT64`, 13: `STRING`, 14: `INT`}

// String returns the string representation of this Type value.
func (i Type) String() string { return enums.String(i, _TypeMap) }

// SetString sets the Type value from its string representation,
// and returns an error if the string is invalid.
func (i *Type) SetString(s string) error { return enums.SetString(i, s, _TypeValueMap, "Type") }

// Int64 returns the Type value as an int64.
func (i Type) Int64() int64 { return int64(i) }

// SetInt64 sets the Type value from an int64.
func (i *Type) SetInt64(in int64) { *i = Type(in) }

// Desc returns the description of the Type value.
func (i Type) Desc() string { return enums.Desc(i, _TypeDescMap) }

// TypeValues returns all possible values for the type Type.
func TypeValues() []Type { return _TypeValues }

// Values returns all possible values for the type Type.
func (i Type) Values() []enums.Enum { return enums.Values(_TypeValues) }

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i Type) MarshalText() ([]byte, error) { return []byte(i.String()), nil }

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *Type) UnmarshalText(text []byte) error { return enums.UnmarshalText(i, text, "Type") }
