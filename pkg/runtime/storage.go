package runtime

import (
	"context"
	"github.com/tetratelabs/wazero/api"
)

// diskr reads up to `size` bytes from persistent storage into the pointer
// `dest`.
func (rt *Runtime) diskr(ctx context.Context, mod api.Module, stack []uint64) {
	/* dest, size */ _, _ = int32(stack[0]), int32(stack[1])
	bytesRead := 0
	stack[0] = uint64(bytesRead)
}

// diskw writes up to `size` bytes from the pointer `src` into persistent
// storage.
func (rt *Runtime) diskw(ctx context.Context, mod api.Module, stack []uint64) {
	/* src, size */ _, _ = int32(stack[0]), int32(stack[1])
	bytesWritten := 0
	stack[0] = uint64(bytesWritten)
}
