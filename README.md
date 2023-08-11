# tg-media

This package provides basic functionality for image/video conversion, to simplify work with `.HEIC`, `.webp`, `.webm`, `.h265`.

Depends on `convert` and `ffmpeg`.

## Example

```go
package main

import (
	tgmedia "github.com/ya-meh/tg-media"
	"log"
	"os"
)

func main() {
	err := tgmedia.PhotoTo("image.heic", "image.jpg")
	if err != nil { log.Fatal(err) }

	temporaryPic, err := tgmedia.Video("input_video.h265.mp4")
	if err != nil { log.Fatal(err) }
	defer os.Remove(temporaryPic)
	
	err = tgmedia.VideoTo("video.webm", "video.mp4")
	if err != nil { log.Fatal(err) }

	temporaryVideo, err := tgmedia.Video("input_video.h265.mp4")
	if err != nil { log.Fatal(err) }
	defer os.Remove(temporaryVideo)

	// your wonderful code...
}
```