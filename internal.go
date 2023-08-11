package tgmedia

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"
)

var (
	ErrorUnimplemented  = errors.New("unsupported for now, make a PR, pls")
	ErrorConvertedBadly = errors.New("converted, but non sendable still")
	ErrorBrokenData     = errors.New("given photo/video is broken")
)

func wrapExecError(output []byte, err error) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s\nstdout:\t%s", err.Error(), string(output))
}

func mktmpFile(pattern string) (string, error) {
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		_ = os.Remove(file.Name())
		return "", err
	}
	if err = file.Close(); err != nil {
		_ = os.Remove(file.Name())
		return "", err
	}
	return file.Name(), nil
}

type fileMetadata struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
	Format struct {
		Filename string `json:"filename"`
		Duration string `json:"duration"`
	} `json:"format"`
}

func getFileMetadata(filename string) (*fileMetadata, error) {
	output, err := exec.Command(FfprobePath, "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height",
		"-of", "json", "-show_format", filename).Output()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\n%s", err, string(output)))
	}

	var metadata fileMetadata
	err = json.Unmarshal(output, &metadata)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("%v\n%s", err, string(output)))
	}

	return &metadata, nil
}

func formatDuration(d time.Duration) string {
	trailingZeros := func(d time.Duration, zeros int) string {
		num := int64(d)
		s := fmt.Sprintf("%d", num)
		for len(s) < zeros {
			s = "0" + s
		}
		return s
	}

	return fmt.Sprintf("%s:%s:%s.%s",
		trailingZeros(d/time.Hour%24, 2), trailingZeros(d/time.Minute%60, 2),
		trailingZeros(d/time.Second%60, 2), trailingZeros(d/time.Millisecond%1000, 3))
}

func contains(array []string, value string) bool {
	for _, el := range array {
		if el == value {
			return true
		}
	}
	return false
}
