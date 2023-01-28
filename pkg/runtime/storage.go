package runtime

import (
	"context"
	"io"
	"os"

	"github.com/tetratelabs/wazero/api"
)

type Storage struct {
	Data     []byte
	Filename string
}

func NewStorage(filename string) *Storage {
	f, err := os.Open(filename)
	if err != nil {
		return &Storage{
			Data:     make([]byte, 0),
			Filename: filename,
		}
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return &Storage{
			Data:     make([]byte, 0),
			Filename: filename,
		}
	}

	return &Storage{
		Data:     data,
		Filename: filename,
	}
}

func (s *Storage) Read(p []byte) (n int, err error) {
	n = copy(p, s.Data)

	return n, nil
}

func (s *Storage) Write(p []byte) (n int, err error) {
	s.Data = make([]byte, len(p))
	n = copy(s.Data, p)

	return n, nil
}

func (s *Storage) Close() error {
	if len(s.Data) == 0 {
		return nil
	}

	f, err := os.Create(s.Filename)
	if err != nil {
		return err
	}

	_, err = f.Write(s.Data)
	if err != nil {
		return err
	}

	return f.Close()
}

// diskr reads up to `size` bytes from persistent storage into the pointer
// `dest`.
func (rt *Runtime) diskr(ctx context.Context, mod api.Module, stack []uint64) {
	var (
		dest = api.DecodeI32(stack[0])
		size = api.DecodeI32(stack[1])
		data = make([]byte, size)
	)

	if rt.cart == nil || rt.cart.Memory() == nil {
		stack[0] = 0
		return
	}

	if size > 1024 {
		size = 1024
	}

	n, err := rt.Storage.Read(data)
	if err != nil || int32(n) != size {
		stack[0] = 0
		return
	}

	data = data[:n]

	ok := rt.cart.Memory().Write(uint32(dest), data)
	if !ok {
		stack[0] = 0
		return
	}

	stack[0] = uint64(n)
}

// diskw writes up to `size` bytes from the pointer `src` into persistent
// storage.
func (rt *Runtime) diskw(ctx context.Context, mod api.Module, stack []uint64) {
	var (
		src  = api.DecodeI32(stack[0])
		size = api.DecodeI32(stack[1])
	)

	if rt.cart == nil || rt.cart.Memory() == nil {
		stack[0] = 0
		return
	}

	if size > 1024 {
		size = 1024
	}

	data, ok := rt.cart.Memory().Read(uint32(src), uint32(size))
	if !ok {
		stack[0] = 0
		return
	}

	m, err := rt.Storage.Write(data)
	if err != nil || int32(m) != size {
		stack[0] = 0
		return
	}

	stack[0] = uint64(m)
}
