package tgmedia

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGeneratePreview(t *testing.T) {
	filename, err := GeneratePreview("testdata/dontworrypolin.webm")
	require.NoError(t, err)
	defer os.Remove(filename)

	stat, err := PhotoStats(filename)
	require.NoError(t, err)
	require.True(t, stat.IsSendable())
}

func TestEncodingsAndConversion(t *testing.T) {
	file, err := mktmpFile("tgmedialib_test_convert_video_*.mp4")
	require.NoError(t, err)
	defer os.Remove(file)

	testfile := "testdata/dontworrypolin.webm" // vp9 -> h264

	supported, err := IsVideoEncodingSupported(testfile)
	require.NoError(t, err)
	require.False(t, supported)

	isEncoded, err := IsVideoEncodedWith(testfile, "vp9")
	require.NoError(t, err)
	require.True(t, isEncoded)

	require.NoError(t, VideoTo(testfile, file))

	supported, err = IsVideoEncodingSupported(file)
	require.NoError(t, err)
	require.True(t, supported)

	//require.NoError(t, exec.Command("vlc", file).Run()) // could be useful, idk
}
