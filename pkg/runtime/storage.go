package runtime

import (
	"context"

	"github.com/tetratelabs/wazero/api"
)

type Storage struct {
	Data []byte
}

func (s *Storage) Read(p []byte) (n int, err error) {
	n = copy(p, s.Data[0:])
	return
}

func (s *Storage) Write(p []byte) (n int, err error) {
	n = copy(s.Data[0:], p)
	return
}

// diskr reads up to `size` bytes from persistent storage into the pointer
// `dest`.
func (rt *Runtime) diskr(ctx context.Context, mod api.Module, stack []uint64) {
	dest, size := int32(stack[0]), int32(stack[1])

	data, ok := rt.cart.Memory().Read(uint32(dest), uint32(dest+size))
	if ok {
		rt.Storage.Write(data)
	}

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
