package tgmedia

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestVerbosePhotoStats(t *testing.T) {
	stat, err := PhotoStats("testdata/Canon_40D.jpg")
	require.NoError(t, err)
	require.True(t, stat.IsSendable())

	stat, err = PhotoStats("testdata/Arbitro.tiff")
	require.NoError(t, err)
	require.False(t, stat.IsSendable())

	stat, err = PhotoStats("testdata/samplefilehub.heif")
	require.NoError(t, err)
	require.False(t, stat.IsSendable())
}

func TestPhoto(t *testing.T) {
	filename, err := Photo("testdata/Canon_40D.jpg")
	require.NoError(t, err)
	require.Equal(t, "testdata/Canon_40D.jpg", filename)

	filename, err = Photo("testdata/samplefilehub.heif")
	require.NoError(t, err)
	defer os.Remove(filename)
	require.NotEqual(t, "testdata/samplefilehub.heif", filename)
}
