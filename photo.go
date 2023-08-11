// Package tg_media provides utility for converting photos to tg suitable formats.
//
//	The photo must be at most 10 MB in size.
//	The photo's width and height must not exceed 10000 in total.
//	Width and height ratio must be at most 20.
//	https://core.telegram.org/bots/api#sendphoto
package tgmedia

import (
	"os"
	"strings"
)

var (
	VarPathConvert = "convert"
	VarJpegQuality = "95"
)

var ApiSupportedImageFormats = []string{"jpg", "png", "jpeg"}

func IsPhotoFormatSupported(filename string) bool {
	f := strings.ToLower(filename)
	for _, format := range ApiSupportedImageFormats {
		if strings.HasSuffix(f, format) {
			return true
		}
	}
	return false
}

func Photo(filename string) (string, error) {
	stat, err := PhotoStats(filename)
	if err != nil {
		return filename, err
	}
	if stat.IsSendable() {
		return filename, nil
	}

	converted, err := mktmpFile("tgmedialib_photo_*.jpg")
	if err != nil {
		return filename, err
	}
	if err = stat.Convert(converted); err != nil {
		_ = os.Remove(converted)
		return filename, err
	}

	return converted, nil
}

func PhotoTo(filename, destination string) error {
	tmp, err := Photo(filename)
	if err != nil {
		return err
	}

	if err = os.Rename(tmp, destination); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	return nil
}
