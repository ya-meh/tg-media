package tgmedia

import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"os/exec"
)

var NotImageError = errors.New("not an image provided")

type Stat struct {
	Name     string `json:"name"`
	FileSize int64  `json:"fileSize"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`

	Format string `json:"format"`
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
	if stat == nil || stat.FileSize == 0 || !IsPhoto(stat.Name) {
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
		if code == CodeNonSendableAspectSize || width == 0 && height == 0 {
			width = 4096
			height = 4096
		}

		stdout, err := exec.Command(VarPathConvert, stat.Name,
			"-size", fmt.Sprintf("%dx%d", width, height),
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
	if !IsPhoto(filename) {
		return nil, NotImageError
	}

	if !IsPhotoFormatSupported(filename) {
		stat, err := os.Stat(filename)
		if err != nil {
			return nil, err
		}

		return &Stat{
			Name:     filename,
			FileSize: stat.Size(),
			Width:    0,
			Height:   0,
			Format:   "",
		}, nil
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		return nil, err
	}

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return &Stat{
		Name:     filename,
		FileSize: stat.Size(),
		Width:    cfg.Width,
		Height:   cfg.Height,
		Format:   format,
	}, nil
}
