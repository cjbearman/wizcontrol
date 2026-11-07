package calendar

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/component/scene"
)

var RandomFrequencyMinutes = 60
var RotationFrequencyMinutes = 60

type SceneSelector interface {
	GetScene(dayNumber int) []*scene.Scene
}

type AllSceneSelector struct {
	scenes []*scene.Scene
}

func (f *AllSceneSelector) GetScene(dayNumber int) []*scene.Scene {
	return f.scenes
}

type RandomSceneSelector struct {
	scenes    []*scene.Scene
	Frequency time.Duration
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
