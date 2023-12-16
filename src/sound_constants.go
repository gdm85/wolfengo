package main

// Sound names used as keys for res/sounds/<name>.wav
// Variants are named <base><1..n>.wav (e.g. monster_pain1.wav … monster_pain3.wav).
const (
	SoundGunshot          = "gunshot"
	SoundPlayerPain       = "player_pain"
	SoundMonsterPainBase  = "monster_pain" // 3 variants: monster_pain1/2/3.wav
	SoundMonsterPainCount = 3
	SoundMonsterDeath     = "monster_death"
	SoundMonsterAlert     = "monster_alert"
	SoundMonsterShoot     = SoundGunshot
	SoundMedkitPickup     = "medkit_pickup"
	SoundDoorOpen         = "door_open"
	SoundDoorClose        = "door_close"
	SoundTeleport         = "teleport"
)
