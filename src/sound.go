//go:build !nosound

package main

/*
#cgo LDFLAGS: -lopenal
#include <AL/al.h>
#include <AL/alc.h>
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"unsafe"
)

const audioSourcePoolSize = 16

var audio struct {
	device  *C.ALCdevice
	context *C.ALCcontext
	buffers map[string]C.ALuint
	sources [audioSourcePoolSize]C.ALuint
}

func initAudio() error {
	audio.device = C.alcOpenDevice(nil)
	if audio.device == nil {
		return fmt.Errorf("alcOpenDevice failed")
	}

	audio.context = C.alcCreateContext(audio.device, nil)
	if audio.context == nil {
		C.alcCloseDevice(audio.device)
		return fmt.Errorf("alcCreateContext failed")
	}

	C.alcMakeContextCurrent(audio.context)
	audio.buffers = make(map[string]C.ALuint)
	C.alGenSources(audioSourcePoolSize, &audio.sources[0])

	// Distance attenuation: volume falls off naturally with 1/distance
	C.alDistanceModel(C.AL_INVERSE_DISTANCE_CLAMPED)

	return nil
}

func shutdownAudio() {
	if audio.context == nil {
		return
	}
	for _, src := range audio.sources {
		C.alSourceStop(src)
	}
	C.alDeleteSources(audioSourcePoolSize, &audio.sources[0])
	for _, buf := range audio.buffers {
		C.alDeleteBuffers(1, &buf)
	}
	C.alcMakeContextCurrent(nil)
	C.alcDestroyContext(audio.context)
	C.alcCloseDevice(audio.device)
	audio.context = nil
}

// updateAudioListener must be called once per frame with the player's camera state.
func updateAudioListener(pos, forward, up Vector3f) {
	C.alListener3f(C.AL_POSITION, C.ALfloat(pos.X), C.ALfloat(pos.Y), C.ALfloat(pos.Z))
	// AL_ORIENTATION takes 6 floats: forward (at) then up
	orientation := [6]C.ALfloat{
		C.ALfloat(forward.X), C.ALfloat(forward.Y), C.ALfloat(forward.Z),
		C.ALfloat(up.X), C.ALfloat(up.Y), C.ALfloat(up.Z),
	}
	C.alListenerfv(C.AL_ORIENTATION, &orientation[0])
}

// playSound plays a non-positional (HUD-level) sound at full volume.
func playSound(name string) {
	playSoundAt(name, Vector3f{}, false)
}

// playSound3D plays a sound anchored at a world-space position.
// Volume and panning are attenuated by distance from the listener.
func playSound3D(name string, pos Vector3f) {
	playSoundAt(name, pos, true)
}

func playSoundAt(name string, pos Vector3f, positional bool) {
	if audio.buffers == nil {
		return
	}
	buf, err := loadBuffer(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "audio:", err)
		return
	}
	src, ok := acquireSource()
	if !ok {
		return // all sources busy; drop this sound
	}
	configureSource(src, buf, pos, positional)
	C.alSourcePlay(src)
}

// monsterSource is a dedicated OpenAL source owned by a single monster.
// Using a per-monster source means AL_SOURCE_STATE always reflects exactly
// what that monster is currently emitting, with no interference from other sounds.
type monsterSource struct{ id C.ALuint }

func newMonsterSource() monsterSource {
	if audio.context == nil {
		return monsterSource{}
	}
	var src C.ALuint
	C.alGenSources(1, &src)
	return monsterSource{id: src}
}

// play stops any current sound on this source and immediately plays name.
func (ms monsterSource) play(name string, pos Vector3f) {
	if soundMuted || audio.buffers == nil || ms.id == 0 {
		return
	}
	buf, err := loadBuffer(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "audio:", err)
		return
	}
	C.alSourceStop(ms.id)
	configureSource(ms.id, buf, pos, true)
	C.alSourcePlay(ms.id)
}

// playIfFree plays name only when this source is not already playing.
func (ms monsterSource) playIfFree(name string, pos Vector3f) {
	if soundMuted || audio.buffers == nil || ms.id == 0 {
		return
	}
	var state C.ALint
	C.alGetSourcei(ms.id, C.AL_SOURCE_STATE, &state)
	if state == C.AL_PLAYING {
		return
	}
	buf, err := loadBuffer(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, "audio:", err)
		return
	}
	configureSource(ms.id, buf, pos, true)
	C.alSourcePlay(ms.id)
}

// free releases the OpenAL source back to the driver.
func (ms monsterSource) free() {
	if ms.id == 0 {
		return
	}
	C.alSourceStop(ms.id)
	C.alDeleteSources(1, &ms.id)
}

func configureSource(src C.ALuint, buf C.ALuint, pos Vector3f, positional bool) {
	C.alSourcei(src, C.AL_BUFFER, C.ALint(buf))
	if positional {
		C.alSource3f(src, C.AL_POSITION, C.ALfloat(pos.X), C.ALfloat(pos.Y), C.ALfloat(pos.Z))
		C.alSourcei(src, C.AL_SOURCE_RELATIVE, C.AL_FALSE)
		C.alSourcef(src, C.AL_REFERENCE_DISTANCE, 1.5) // full volume within ~1.5 units
		C.alSourcef(src, C.AL_MAX_DISTANCE, 15.0)
		C.alSourcef(src, C.AL_ROLLOFF_FACTOR, 1.0)
	} else {
		C.alSourcei(src, C.AL_SOURCE_RELATIVE, C.AL_TRUE)
		C.alSource3f(src, C.AL_POSITION, 0, 0, 0)
	}
}

// acquireSource returns a source from the pool that is not currently playing.
func acquireSource() (C.ALuint, bool) {
	for _, src := range audio.sources {
		var state C.ALint
		C.alGetSourcei(src, C.AL_SOURCE_STATE, &state)
		if state != C.AL_PLAYING {
			return src, true
		}
	}
	return 0, false
}

// loadBuffer loads and caches a WAV file as an OpenAL buffer.
// Missing files are non-fatal: the error is returned and the caller logs it.
func loadBuffer(name string) (C.ALuint, error) {
	if buf, ok := audio.buffers[name]; ok {
		return buf, nil
	}
	path := "./res/sounds/" + name + ".wav"
	data, format, sampleRate, err := loadWAV(path)
	if err != nil {
		return 0, err
	}
	var buf C.ALuint
	C.alGenBuffers(1, &buf)
	C.alBufferData(buf, format, unsafe.Pointer(&data[0]), C.ALsizei(len(data)), C.ALsizei(sampleRate))
	audio.buffers[name] = buf
	return buf, nil
}

// -- WAV loader (PCM only: mono/stereo, 8 or 16 bit) -------------------------

type wavFmtChunk struct {
	AudioFormat   uint16
	NumChannels   uint16
	SampleRate    uint32
	ByteRate      uint32
	BlockAlign    uint16
	BitsPerSample uint16
}

func loadWAV(path string) (data []byte, format C.ALenum, sampleRate int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	// RIFF header: "RIFF" <4-byte size> "WAVE"
	var header [12]byte
	if _, err = io.ReadFull(f, header[:]); err != nil {
		return
	}
	if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
		err = fmt.Errorf("not a WAV file: %s", path)
		return
	}

	var fmtChunk wavFmtChunk

	for {
		var id [4]byte
		if _, err = io.ReadFull(f, id[:]); err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				err = nil
			}
			break
		}
		var size uint32
		if err = binary.Read(f, binary.LittleEndian, &size); err != nil {
			return
		}
		switch string(id[:]) {
		case "fmt ":
			if err = binary.Read(f, binary.LittleEndian, &fmtChunk); err != nil {
				return
			}
			if size > 16 {
				f.Seek(int64(size-16), io.SeekCurrent)
			}
		case "data":
			data = make([]byte, size)
			if _, err = io.ReadFull(f, data); err != nil {
				return
			}
		default:
			f.Seek(int64(size), io.SeekCurrent)
		}
	}

	if len(data) == 0 {
		err = fmt.Errorf("WAV has no data: %s", path)
		return
	}

	switch {
	case fmtChunk.NumChannels == 1 && fmtChunk.BitsPerSample == 8:
		format = C.AL_FORMAT_MONO8
	case fmtChunk.NumChannels == 1 && fmtChunk.BitsPerSample == 16:
		format = C.AL_FORMAT_MONO16
	case fmtChunk.NumChannels == 2 && fmtChunk.BitsPerSample == 8:
		format = C.AL_FORMAT_STEREO8
	case fmtChunk.NumChannels == 2 && fmtChunk.BitsPerSample == 16:
		format = C.AL_FORMAT_STEREO16
	default:
		err = fmt.Errorf("unsupported WAV format (%d ch, %d bit): %s",
			fmtChunk.NumChannels, fmtChunk.BitsPerSample, path)
	}
	sampleRate = int(fmtChunk.SampleRate)
	return
}
