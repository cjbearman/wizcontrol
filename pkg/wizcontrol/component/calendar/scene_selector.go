package calendar

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
)

// RandomFrequencyMinutes is the default frequency that we will randomly rotate scenes, unless overridden
var RandomFrequencyMinutes = 60

// RotationFrequencyMinutes is the default frequency that we will rotate scenes, unless overridden
var RotationFrequencyMinutes = 60

// SceneSelector defines the interface for selecting scenes based on the day number.
type SceneSelector interface {
	GetScene(dayNumber int) []*scene.Scene
	FrequencyStr() string
	ModeStr() string
}

// AllSceneSelector returns all scenes
type AllSceneSelector struct {
	scenes []*scene.Scene
}

// GetScene returns all scenes, regardless of the day number.
func (f *AllSceneSelector) GetScene(dayNumber int) []*scene.Scene {
	return f.scenes
}

// ModeStr returns the mode string for the AllSceneSelector. It will always return SceneModeAll.
func (f *AllSceneSelector) ModeStr() string {
	return SceneModeAll
}

// FrequencyStr returns the frequency string for the AllSceneSelector. It will always return an empty string.
func (f *AllSceneSelector) FrequencyStr() string {
	return ""
}

// RandomSceneSelector selects scenes randomly based on a defined frequency.
type RandomSceneSelector struct {
	scenes    []*scene.Scene
	Frequency time.Duration
}

// ModeStr returns the mode string for the RandomSceneSelector. It will always return SceneModeRandom.
func (f *RandomSceneSelector) ModeStr() string {
	return SceneModeRandom
}

// FrequencyStr returns the frequency string for the RandomSceneSelector.
func (f *RandomSceneSelector) FrequencyStr() string {
	return f.Frequency.String()
}

// GetScene returns a pseudo-random scene based on the current time and a defined
// change frequency defined by RandomFrequencyMinutes
// It returns a psuedo-random scene that remains constant for the given frequency minutes
// but will then (likely) change for the next period
func (f *RandomSceneSelector) GetScene(dayNumber int) []*scene.Scene {
	// Number of minutes since epoch
	minutesSinceEpoch := time.Now().In(common.Location).Unix() / 60

	// How often, in seconds, to change the scene
	frequencyMins := f.Frequency.Minutes()

	// Divide this by our frequency of desired change
	frequencyOffset := minutesSinceEpoch / int64(frequencyMins)

	// This frequency offset will basically change every X minutes
	// We'll add that into a hashable but otherwise consistent value and hash it to get a pseudo-random selection
	sceneHashable := fmt.Sprintf("abc123%d", frequencyOffset)

	hash := sha256.Sum256([]byte(sceneHashable))
	index := (int(hash[0]) + int(hash[1]) + int(hash[2])) % len(f.scenes)
	return []*scene.Scene{f.scenes[index]}
}

// RotateSceneSelector selects scenes in a rotating order based on a defined frequency.
type RotateSceneSelector struct {
	scenes    []*scene.Scene
	Frequency time.Duration
}

// GetScene returns scenes in order, changing on the given frequency
func (f *RotateSceneSelector) GetScene(dayNumber int) []*scene.Scene {
	// Number of minutes since epoch
	minutesSinceEpoch := time.Now().In(common.Location).Unix() / 60

	// How often, in minutes, to change the scene
	frequencyMins := f.Frequency.Minutes()

	// Divide this by our frequency of desired change
	frequencyOffset := minutesSinceEpoch / int64(frequencyMins)

	// Use the frequency offset to determine the index
	index := int(frequencyOffset) % len(f.scenes)

	return []*scene.Scene{f.scenes[index]}
}

// ModeStr returns the mode string for the RotateSceneSelector. It will always return SceneModeRotate.
func (f *RotateSceneSelector) ModeStr() string {
	return SceneModeRotate
}

// FrequencyStr returns the frequency string for the RotateSceneSelector.
func (f *RotateSceneSelector) FrequencyStr() string {
	return f.Frequency.String()
}
