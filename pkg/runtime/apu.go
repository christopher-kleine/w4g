package runtime

type APU struct{}

func (apu *APU) Tone(frequency, duration, volume, flags int32) {
	//log.Printf("tone: %d / %d / %d / %d", frequency, duration, volume, flags)
}
