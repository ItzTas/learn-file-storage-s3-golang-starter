package videos

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

type AspectRatio string

const (
	AspectOther AspectRatio = "other"
	Aspect169   AspectRatio = "16:9"
	Aspect916   AspectRatio = "9:16"
)

func GetVideoAspectRatio(filepath string) (AspectRatio, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_streams",
		filepath,
	)
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return AspectOther, err
	}
	var vm VideoMetadata
	if err := json.Unmarshal(buf.Bytes(), &vm); err != nil {
		return AspectOther, err
	}
	return calculateAspectRatio(vm), nil
}

func calculateAspectRatio(vm VideoMetadata) AspectRatio {
	if len(vm.Streams) == 0 {
		return AspectOther
	}

	width := float64(vm.Streams[0].Width)
	height := float64(vm.Streams[0].Height)
	ratio := width / height

	if closeEnough(ratio, 16.0/9.0) {
		return Aspect169
	}
	if closeEnough(ratio, 9.0/16.0) {
		return Aspect916
	}
	return AspectOther
}

func closeEnough(a, b float64) bool {
	const epsilon = 0.01
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff < epsilon
}
