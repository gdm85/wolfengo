package main

import "flag"

var (
	cfgDebugGL       bool
	cfgDebugHitboxes bool
	cfgDebugMonsters bool
	cfgPrintFPS      bool
	cfgNoSound       bool
	cfgLevelFile     string
)

func init() {
	flag.BoolVar(&cfgDebugGL, "debug-gl", false, "enable GL debug callbacks")
	flag.BoolVar(&cfgDebugHitboxes, "debug-hitboxes", false, "draw monster collision hitboxes")
	flag.BoolVar(&cfgDebugMonsters, "debug-monsters", false, "log monster alert and shoot events")
	flag.BoolVar(&cfgPrintFPS, "fps", false, "show FPS counter in the HUD (top-right)")
	flag.BoolVar(&cfgNoSound, "nosound", false, "disable sound at startup (no-op if built with -tags nosound)")
	flag.StringVar(&cfgLevelFile, "level", "", "path to a .map file to load as the first level")
}
