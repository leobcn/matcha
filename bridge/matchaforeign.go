// +build matcha

package bridge

// Go support functions for Objective-C. Note that this
// file is copied into and compiled with the generated
// bindings.

/*
#include <stdbool.h>
#include <stdint.h>
#include <stdlib.h>
#include "matchaforeign.h"
*/
import "C"

import (
	"fmt"
	"reflect"
	"runtime"
)

//export TestFunc
func TestFunc() {
	a := Bool(true)
	b := Int64(1234)
	c := Float64(1.234)
	d := String("abc")
	e := Bytes([]byte("def123"))

	fmt.Println("blah", a.ToBool(), b.ToInt64(), c.ToFloat64(), d.ToString(), string(e.ToBytes()), "~")

	arr := Array(a, b, c, d, e)
	fmt.Println("done")
	arr2 := arr.ToArray()

	fmt.Println("blah2", arr2[0].ToBool(), arr2[1].ToInt64(), arr2[2].ToFloat64(), arr2[3].ToString(), string(arr2[4].ToBytes()), "~")

	bridge := Bridge("a")
	fmt.Println("bridge", bridge)
}

type Value struct {
	ref int64
}

func newValue(ref C.ObjcRef) *Value {
	v := &Value{ref: int64(ref)}
	if ref != 0 {
		runtime.SetFinalizer(v, func(a *Value) {
			C.MatchaUntrackObjc(a._ref())
		})
	}
	return v
}

func (v *Value) _ref() C.ObjcRef {
	return C.ObjcRef(v.ref)
}

func Bridge(a string) *Value {
	cstr := cString(a)
	return newValue(C.MatchaForeignBridge(cstr))
}

func Nil() *Value {
	return newValue(C.ObjcRef(0))
}

func (v *Value) IsNil() bool {
	return v.ref == 0
}

func Bool(v bool) *Value {
	return newValue(C.MatchaObjcBool(C.bool(v)))
}

func (v *Value) ToBool() bool {
	defer runtime.KeepAlive(v)
	return bool(C.MatchaObjcToBool(v._ref()))
}

func Int64(v int64) *Value {
	return newValue(C.MatchaObjcInt64(C.int64_t(v)))
}

func (v *Value) ToInt64() int64 {
	defer runtime.KeepAlive(v)
	return int64(C.MatchaObjcToInt64(v._ref()))
}

func Float64(v float64) *Value {
	return newValue(C.MatchaObjcFloat64(C.double(v)))
}

func (v *Value) ToFloat64() float64 {
	defer runtime.KeepAlive(v)
	return float64(C.MatchaObjcToFloat64(v._ref()))
}

func String(v string) *Value {
	cstr := cString(v)
	return newValue(C.MatchaObjcString(cstr))
}

func (v *Value) ToString() string {
	defer runtime.KeepAlive(v)
	buf := C.MatchaObjcToString(v._ref())
	return goString(buf)
}

func Bytes(v []byte) *Value {
	cbytes := cBytes(v)
	return newValue(C.MatchaObjcBytes(cbytes))
}

func (v *Value) ToBytes() []byte {
	defer runtime.KeepAlive(v)
	buf := C.MatchaObjcToBytes(v._ref())
	return goBytes(buf)
}

func Interface(v interface{}) *Value {
	// Start with a go value.
	// Reflect on it.
	rv := reflect.ValueOf(v)
	// Track it, turning it into a goref.
	ref := matchaGoTrack(rv)
	// Wrap the goref in an foreign object, returning a foreign ref.
	return newValue(C.MatchaObjcGoRef(ref))
}

func (v *Value) ToInterface() interface{} {
	defer runtime.KeepAlive(v)
	// Start with a foreign ref, referring to a foreign value wrapping a go ref.
	// Get the goref.
	ref := C.MatchaObjcToGoRef(v._ref())
	// Get the go object, and unreflect.
	return matchaGoGet(ref).Interface()
}

func Array(a ...*Value) *Value {
	ref := C.MatchaObjcArray(C.int64_t(len(a)))
	array := newValue(ref)
	for idx, i := range a {
		C.MatchaObjcArraySet(ref, i._ref(), C.int64_t(idx))
	}
	return array
}

func (v *Value) ToArray() []*Value {
	defer runtime.KeepAlive(v)
	ref := v._ref()
	length := int64(C.MatchaObjcArrayLen(v._ref()))

	slice := make([]*Value, length)
	for i := int64(0); i < length; i++ {
		slice[i] = newValue(C.MatchaObjcArrayAt(ref, C.int64_t(i)))
	}
	return slice
}

func callSentinel() *Value {
	return newValue(C.MatchaObjcCallSentinel())
}

// Call accepts `nil` in its variadic arguments
func (v *Value) Call(s string, args ...*Value) *Value {
	defer runtime.KeepAlive(v)

	if runtime.GOOS == "darwin" {
		// Can't pass nil through NSArray so put a sentinel in.
		for i, elem := range args {
			if elem == nil || elem.IsNil() {
				args[i] = callSentinel()
			}
		}
	}

	argsValue := Nil()
	if len(args) > 0 {
		argsValue = Array(args...)
		defer runtime.KeepAlive(argsValue)
	}
	return newValue(C.MatchaObjcCall(v._ref(), cString(s), argsValue._ref()))
}

func cBytes(v []byte) C.CGoBuffer {
	var cstr C.CGoBuffer
	if len(v) == 0 {
		cstr = C.CGoBuffer{}
	} else {
		cstr = C.CGoBuffer{
			ptr: C.CBytes(v),
			len: C.int64_t(len(v)),
		}
	}
	return cstr
}

func cString(v string) C.CGoBuffer {
	var cstr C.CGoBuffer
	if len(v) == 0 {
		cstr = C.CGoBuffer{}
	} else {
		cstr = C.CGoBuffer{
			ptr: C.CBytes([]byte(v)),
			len: C.int64_t(len(v)),
		}
	}
	return cstr
}

func goString(buf C.CGoBuffer) string {
	defer C.free(buf.ptr)
	str := C.GoBytes(buf.ptr, C.int(buf.len))
	return string(str)
}

func goBytes(buf C.CGoBuffer) []byte {
	defer C.free(buf.ptr)
	return C.GoBytes(buf.ptr, C.int(buf.len))
}
