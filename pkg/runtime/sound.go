package runtime

import "context"

// tone plays a sound tone.
func (rt *Runtime) tone(_ context.Context, params []uint64) {
	frequency := int32(params[0])
	duration := int32(params[1])
	volume := int32(params[2])
	flags := int32(params[3])

	rt.APU.Tone(frequency, duration, volume, flags)
}
