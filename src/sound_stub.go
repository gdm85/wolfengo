//go:build nosound

package main

var soundMuted bool

func setAudioMuted(muted bool) { soundMuted = muted }
func toggleAudioMute()         { soundMuted = !soundMuted }

func initAudio() error                              { return nil }
func shutdownAudio()                                {}
func updateAudioListener(pos, forward, up Vector3f) {}
func playSound(name string)             {}
func playSound3D(name string, pos Vector3f) {}

type monsterSource struct{}

func newMonsterSource() monsterSource               { return monsterSource{} }
func (ms monsterSource) play(name string, pos Vector3f)      {}
func (ms monsterSource) playIfFree(name string, pos Vector3f) {}
func (ms monsterSource) free()                               {}
