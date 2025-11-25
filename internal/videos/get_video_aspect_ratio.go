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

	width := vm.Streams[0].Width
	height := vm.Streams[0].Height

	if width*9 == height*16 {
		return Aspect169
	}
	if width*16 == height*9 {
		return Aspect916
	}
	return AspectOther
}
