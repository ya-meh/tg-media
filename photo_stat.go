package tgmedia

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type convertJsonWrapper []struct {
	Version string
	Image   StatVerbose
}

type StatVerbose struct {
	Name              string `json:"name"`
	BaseName          string `json:"baseName"`
	Format            string `json:"format"`
	FormatDescription string `json:"formatDescription"`
	MimeType          string `json:"mimeType"`
	Geometry          struct {
		Width  int `json:"width"`
		Height int `json:"height"`
		X      int `json:"x"`
		Y      int `json:"y"`
	} `json:"geometry"`
	Resolution struct {
		X int `json:"x"`
		Y int `json:"y"`
	} `json:"resolution"`
	Colorspace  string `json:"colorspace"`
	Pixels      int    `json:"pixels"`
	Quality     int    `json:"quality"`
	Orientation string `json:"orientation"`
	Properties  struct {
		DateCreate    time.Time `json:"date:create"`
		DateModify    time.Time `json:"date:modify"`
		DateTimestamp time.Time `json:"date:timestamp"`
	} `json:"properties"`
	FileSize int64 `json:"fileSize__"`
}

func (stat *StatVerbose) ToStat() *Stat {
	if stat == nil {
		return nil
	}
	return &Stat{
		FileSize: stat.FileSize,
		Width:    stat.Geometry.Width,
		Height:   stat.Geometry.Height,
		Format:   stat.Format,
		MimeType: stat.MimeType,
		Name:     stat.Name,
	}
}

func VerbosePhotoStats(convertPath, filepath string) (*StatVerbose, error) {
	output, err := exec.Command(convertPath, filepath, "json:-").Output()
	if err != nil {
		return nil, wrapExecError(output, err)
	}
	stat, err := os.Stat(filepath)
	if err != nil {
		return nil, err
	}

	var wrapped convertJsonWrapper
	if err = json.Unmarshal(output, &wrapped); err != nil {
		return nil, err
	}
	if len(wrapped) < 1 {
		return nil, errors.New("stat was not found")
	}

	wrapped[0].Image.FileSize = stat.Size()

	return &wrapped[0].Image, nil
}

type Stat struct {
	Name     string `json:"name"`
	FileSize int64  `json:"fileSize"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`

	Format   string `json:"format"`
	MimeType string `json:"mimeType"`
}

const (
	CodeSendable          = iota
	CodeNonSendableBroken = iota
	CodeNonSendableAspectSize
	CodeNonSendableAspectRatio
	CodeNonSendableFormat
	CodeNonSendableFileSize
)

func (stat *Stat) SendableCode() int {
	if stat == nil || stat.Width == 0 || stat.Height == 0 || stat.FileSize == 0 || !strings.HasPrefix(strings.ToLower(stat.MimeType), "image") {
		return CodeNonSendableBroken
	}

	if stat.Width+stat.Height > 9999 {
		return CodeNonSendableAspectSize
	}

	aspectRatio := float64(stat.Width) / float64(stat.Height)
	if aspectRatio >= 20 || 1/aspectRatio >= 20 {
		return CodeNonSendableAspectRatio
	}

	if stat.FileSize >= 10*1024*1024 {
		return CodeNonSendableFileSize
	}

	if !IsPhotoFormatSupported(stat.Name) {
		return CodeNonSendableFormat
	}

	return CodeSendable
}

func (stat *Stat) IsSendable() bool {
	return stat.SendableCode() == CodeSendable
}

func (stat *Stat) Convert(destination string) error {
	code := stat.SendableCode()
	switch code {
	case CodeSendable:
		return nil
	case CodeNonSendableBroken:
		return ErrorBrokenData
	case CodeNonSendableAspectRatio:
		return ErrorUnimplemented
	default:
		width, height := stat.Width, stat.Height
		if code == CodeNonSendableAspectSize {
			coef := 10000 / float64(width+height)
			// in case of precision error
			width = int(float64(width)*coef) - 5
			height = int(float64(width)*coef) - 5
		}

		stdout, err := exec.Command(VarPathConvert, stat.Name,
			"-resize", fmt.Sprintf("%dx%d", width, height),
			"-quality", VarJpegQuality,
			destination,
		).Output()
		if err != nil {
			return wrapExecError(stdout, err)
		}

		stat, err := PhotoStats(destination)
		if err != nil {
			return err
		}
		if !stat.IsSendable() {
			return ErrorConvertedBadly
		}

		return nil
	}
}

func PhotoStats(filename string) (*Stat, error) {
	verbose, err := VerbosePhotoStats(VarPathConvert, filename)
	return verbose.ToStat(), err
}
