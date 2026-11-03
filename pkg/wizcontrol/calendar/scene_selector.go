package calendar

import (
	"crypto/sha256"
	"fmt"
)

type SceneSelector interface {
	GetScene(dayNumber int) string
}

type FixedSceneSelector struct {
	scene string
}

func (f *FixedSceneSelector) GetScene(dayNumber int) string {
	return f.scene
}

type RandomSceneSelector struct {
	scenes []string
}

func (f *RandomSceneSelector) GetScene(dayNumber int) string {
	// Not really random, hash based so we get the same seen every day
	sceneHashable := fmt.Sprintf("abc123%d", dayNumber)
	hash := sha256.Sum256([]byte(sceneHashable))
	index := (int(hash[0]) + int(hash[1]) + int(hash[2])) % len(f.scenes)
	return f.scenes[index]
}

type RotateSceneSelector struct {
	scenes []string
}

func (f *RotateSceneSelector) GetScene(dayNumber int) string {
	index := dayNumber % len(f.scenes)
	return f.scenes[index]
}
