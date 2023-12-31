package tgmedia

import (
	"errors"
	"fmt"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/matchers"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// I do not use constants on the purpose of more versatile API
var (
	FfmpegPath         = "ffmpeg"
	FfprobePath        = "ffprobe"
	ffmpegConvert      = []string{"-vcodec", "libx264", "-acodec", "aac", "-y", "-preset", "fast"}
	ffprobeGetEncoding = strings.Split("-v error -select_streams v:0 -show_entries stream=codec_name -of default=noprint_wrappers=1:nokey=1", " ")
	TargetEncoding     = "h264"
	SupportedEncodings = []string{"h264", "hevc", "h265"}
)

func IsVideoEncodingSupported(filename string) (bool, error) {
	videoEncoding, err := GetVideoEncoding(filename)
	if err != nil {
		return false, err
	}
	return contains(SupportedEncodings, videoEncoding), nil
}

func IsVideoEncodedWith(filename, encoding string) (bool, error) {
	videoEncoding, err := GetVideoEncoding(filename)
	if err != nil {
		return false, err
	}
	return strings.ToLower(videoEncoding) == strings.ToLower(encoding), nil
}

func GetVideoEncoding(filename string) (string, error) {
	args := []string{}
	args = append(args, ffprobeGetEncoding...)
	args = append(args, filename)

	stdout, err := exec.Command(FfprobePath, args...).Output()
	if err != nil {
		return "", wrapExecError(stdout, err)
	}

	return strings.Trim(string(stdout), "\n "), nil
}

var NotVideoProvided = errors.New("video expected")

func IsVideo(filename string) bool {
	buff, err := ReadTop(filename, 512)
	if err != nil {
		return false
	}

	return filetype.IsVideo(buff)
}

func IsFormatSupported(filename string) bool {
	buff, err := ReadTop(filename, 512)
	if err != nil {
		return false
	}

	return matchers.Mp4(buff) || matchers.Mov(buff)
}

func Video(filename string) (string, error) {
	if !IsVideo(filename) {
		return "", NotVideoProvided
	}

	if IsFormatSupported(filename) {
		return filename, nil
	}

	file, err := mktmpFile("tgmedialib_video_*.mp4")
	if err != nil {
		return filename, err
	}

	if err = VideoTo(filename, file); err != nil {
		_ = os.Remove(file)
		return filename, err
	}

	return file, err
}

func VideoTo(source, destination string) error {
	args := []string{}
	args = append(args, "-i", source)
	args = append(args, ffmpegConvert...)
	args = append(args, destination)

	stdout, err := exec.Command(FfmpegPath, args...).Output()
	if err != nil {
		return wrapExecError(stdout, err)
	}

	return nil
}

func GeneratePreviewTo(video, thumbnail string) error {
	metadata, err := getFileMetadata(video)
	if err != nil {
		return err
	}
	bigThumbnail, err := mktmpFile("tgmedialib_video_thumbnail_big_*.jpg")
	defer os.Remove(bigThumbnail)
	if err != nil {
		return err
	}

	duration, _ := strconv.ParseFloat(metadata.Format.Duration, 10)
	output, err := exec.Command(FfmpegPath, "-y", "-i", video,
		"-ss", formatDuration(time.Duration(duration*float64(time.Second/2))),
		"-vframes", "1", bigThumbnail).Output()
	if err != nil {
		return err
	}

	output, err = exec.Command(VarPathConvert, bigThumbnail, "-resize", "320x320", "-quality", "87", thumbnail).Output()
	if err != nil {
		return fmt.Errorf("%v\n%s", err, string(output))
	}

	return nil
}

func GeneratePreview(filename string) (string, error) {
	thumbnail, err := mktmpFile("tgmedialib_video_thumbnail_*.jpg")
	if err != nil {
		return "", err
	}

	if err = GeneratePreviewTo(filename, thumbnail); err != nil {
		return "", err
	}

	return thumbnail, nil
}
