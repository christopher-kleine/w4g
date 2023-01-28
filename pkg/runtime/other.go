package runtime

import (
	"context"
	"log"

	"github.com/tetratelabs/wazero/api"
)

// trace prints a message to the debug console from a *zero-terminated*
// string pointer.
func (rt *Runtime) trace(_ context.Context, mod api.Module, params []uint64) {
	str := int32(params[0])
	/* message */ _ = getString(mod.Memory(), str)
}

// traceUtf8 prints a message to the debug console from a UTF-8 encoded
// input.
func (rt *Runtime) traceUtf8(_ context.Context, mod api.Module, params []uint64) {
	str := int32(params[0])
	byteLength := int32(params[1])

	/* message */
	message := string(readBytes(mod, str, byteLength))
	log.Println(message)
}

// traceUtf16 prints a message to the debug console from a UTF-16 encoded
// input.
func (rt *Runtime) traceUtf16(_ context.Context, mod api.Module, params []uint64) {
	str := int32(params[0])
	byteLength := int32(params[1])

	/* message */
	message := DecodeUTF16(readBytes(mod, str, byteLength))
	log.Println(message)
}

// tracef prints a message to the debug console from the following input:
//
// * %c, %d, and %x expect 32-bit integers.
// * %f expects 64-bit floats.
// * %s expects a *zero-terminated* string pointer.
func (rt *Runtime) tracef(_ context.Context, mod api.Module, params []uint64) {
	/* str := */ _ = int32(params[0])
	/* stack := */ _ = int32(params[1])
}
