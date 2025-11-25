package videos

import "os/exec"

func ProcessVideoForFastStart(filepath string) (string, error) {
	outputFileName := filepath + ".processing"
	cmd := exec.Command(
		"ffmpeg",
		"-i", filepath,
		"-c", "copy",
		"-movflags", "faststart",
		"-f", "mp4",
		outputFileName,
	)
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return outputFileName, nil
}
