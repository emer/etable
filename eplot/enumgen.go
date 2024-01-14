// Code generated by "goki generate"; DO NOT EDIT.

package eplot

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"goki.dev/enums"
)

var _PlotTypesValues = []PlotTypes{0, 1}

// PlotTypesN is the highest valid value
// for type PlotTypes, plus one.
const PlotTypesN PlotTypes = 2

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the enumgen command to generate them again.
func _PlotTypesNoOp() {
	var x [1]struct{}
	_ = x[XY-(0)]
	_ = x[Bar-(1)]
}

var _PlotTypesNameToValueMap = map[string]PlotTypes{
	`XY`:  0,
	`xy`:  0,
	`Bar`: 1,
	`bar`: 1,
}

var _PlotTypesDescMap = map[PlotTypes]string{
	0: `XY is a standard line / point plot`,
	1: `Bar plots vertical bars`,
}

var _PlotTypesMap = map[PlotTypes]string{
	0: `XY`,
	1: `Bar`,
}

// String returns the string representation
// of this PlotTypes value.
func (i PlotTypes) String() string {
	if str, ok := _PlotTypesMap[i]; ok {
		return str
	}
	return strconv.FormatInt(int64(i), 10)
}

// SetString sets the PlotTypes value from its
// string representation, and returns an
// error if the string is invalid.
func (i *PlotTypes) SetString(s string) error {
	if val, ok := _PlotTypesNameToValueMap[s]; ok {
		*i = val
		return nil
	}
	if val, ok := _PlotTypesNameToValueMap[strings.ToLower(s)]; ok {
		*i = val
		return nil
	}
	return errors.New(s + " is not a valid value for type PlotTypes")
}

// Int64 returns the PlotTypes value as an int64.
func (i PlotTypes) Int64() int64 {
	return int64(i)
}

// SetInt64 sets the PlotTypes value from an int64.
func (i *PlotTypes) SetInt64(in int64) {
	*i = PlotTypes(in)
}

// Desc returns the description of the PlotTypes value.
func (i PlotTypes) Desc() string {
	if str, ok := _PlotTypesDescMap[i]; ok {
		return str
	}
	return i.String()
}

// PlotTypesValues returns all possible values
// for the type PlotTypes.
func PlotTypesValues() []PlotTypes {
	return _PlotTypesValues
}

// Values returns all possible values
// for the type PlotTypes.
func (i PlotTypes) Values() []enums.Enum {
	res := make([]enums.Enum, len(_PlotTypesValues))
	for i, d := range _PlotTypesValues {
		res[i] = d
	}
	return res
}

// IsValid returns whether the value is a
// valid option for type PlotTypes.
func (i PlotTypes) IsValid() bool {
	_, ok := _PlotTypesMap[i]
	return ok
}

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i PlotTypes) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *PlotTypes) UnmarshalText(text []byte) error {
	if err := i.SetString(string(text)); err != nil {
		log.Println("PlotTypes.UnmarshalText:", err)
	}
	return nil
}

var _ShapesValues = []Shapes{0, 1, 2, 3, 4, 5, 6, 7}

// ShapesN is the highest valid value
// for type Shapes, plus one.
const ShapesN Shapes = 8

// An "invalid array index" compiler error signifies that the constant values have changed.
// Re-run the enumgen command to generate them again.
func _ShapesNoOp() {
	var x [1]struct{}
	_ = x[Ring-(0)]
	_ = x[Circle-(1)]
	_ = x[Square-(2)]
	_ = x[Box-(3)]
	_ = x[Triangle-(4)]
	_ = x[Pyramid-(5)]
	_ = x[Plus-(6)]
	_ = x[Cross-(7)]
}

var _ShapesNameToValueMap = map[string]Shapes{
	`Ring`:     0,
	`ring`:     0,
	`Circle`:   1,
	`circle`:   1,
	`Square`:   2,
	`square`:   2,
	`Box`:      3,
	`box`:      3,
	`Triangle`: 4,
	`triangle`: 4,
	`Pyramid`:  5,
	`pyramid`:  5,
	`Plus`:     6,
	`plus`:     6,
	`Cross`:    7,
	`cross`:    7,
}

var _ShapesDescMap = map[Shapes]string{
	0: `Ring is the outline of a circle`,
	1: `Circle is a solid circle`,
	2: `Square is the outline of a square`,
	3: `Box is a filled square`,
	4: `Triangle is the outline of a triangle`,
	5: `Pyramid is a filled triangle`,
	6: `Plus is a plus sign`,
	7: `Cross is a big X`,
}

var _ShapesMap = map[Shapes]string{
	0: `Ring`,
	1: `Circle`,
	2: `Square`,
	3: `Box`,
	4: `Triangle`,
	5: `Pyramid`,
	6: `Plus`,
	7: `Cross`,
}

// String returns the string representation
// of this Shapes value.
func (i Shapes) String() string {
	if str, ok := _ShapesMap[i]; ok {
		return str
	}
	return strconv.FormatInt(int64(i), 10)
}

// SetString sets the Shapes value from its
// string representation, and returns an
// error if the string is invalid.
func (i *Shapes) SetString(s string) error {
	if val, ok := _ShapesNameToValueMap[s]; ok {
		*i = val
		return nil
	}
	if val, ok := _ShapesNameToValueMap[strings.ToLower(s)]; ok {
		*i = val
		return nil
	}
	return errors.New(s + " is not a valid value for type Shapes")
}

// Int64 returns the Shapes value as an int64.
func (i Shapes) Int64() int64 {
	return int64(i)
}

// SetInt64 sets the Shapes value from an int64.
func (i *Shapes) SetInt64(in int64) {
	*i = Shapes(in)
}

// Desc returns the description of the Shapes value.
func (i Shapes) Desc() string {
	if str, ok := _ShapesDescMap[i]; ok {
		return str
	}
	return i.String()
}

// ShapesValues returns all possible values
// for the type Shapes.
func ShapesValues() []Shapes {
	return _ShapesValues
}

// Values returns all possible values
// for the type Shapes.
func (i Shapes) Values() []enums.Enum {
	res := make([]enums.Enum, len(_ShapesValues))
	for i, d := range _ShapesValues {
		res[i] = d
	}
	return res
}

// IsValid returns whether the value is a
// valid option for type Shapes.
func (i Shapes) IsValid() bool {
	_, ok := _ShapesMap[i]
	return ok
}

// MarshalText implements the [encoding.TextMarshaler] interface.
func (i Shapes) MarshalText() ([]byte, error) {
	return []byte(i.String()), nil
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
func (i *Shapes) UnmarshalText(text []byte) error {
	if err := i.SetString(string(text)); err != nil {
		log.Println("Shapes.UnmarshalText:", err)
	}
	return nil
}
