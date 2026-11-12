package proof

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/Lil-Strudel/glassact-studios/apps/api/svg"
)

// errUnmirrorableAsset means the design file's format cannot be flipped here
// (a PDF, say), so production has to mirror it by hand.
var errUnmirrorableAsset = errors.New("design asset format cannot be mirrored")

// mirrorDesignAsset flips a design file the way its inlay is mirrored and
// returns the result with its content type.
func mirrorDesignAsset(content []byte, ext string, mirrorX, mirrorY bool) ([]byte, string, error) {
	switch strings.ToLower(ext) {
	case ".svg":
		out, err := svg.Mirror(content, mirrorX, mirrorY)
		return out, "image/svg+xml", err
	case ".png":
		img, err := png.Decode(bytes.NewReader(content))
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode png design: %w", err)
		}
		var out bytes.Buffer
		if err := png.Encode(&out, flipImage(img, mirrorX, mirrorY)); err != nil {
			return nil, "", fmt.Errorf("failed to encode mirrored png design: %w", err)
		}
		return out.Bytes(), "image/png", nil
	case ".jpg", ".jpeg":
		img, err := jpeg.Decode(bytes.NewReader(content))
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode jpeg design: %w", err)
		}
		var out bytes.Buffer
		if err := jpeg.Encode(&out, flipImage(img, mirrorX, mirrorY), &jpeg.Options{Quality: 95}); err != nil {
			return nil, "", fmt.Errorf("failed to encode mirrored jpeg design: %w", err)
		}
		return out.Bytes(), "image/jpeg", nil
	default:
		return nil, "", errUnmirrorableAsset
	}
}

func flipImage(src image.Image, mirrorX, mirrorY bool) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		srcY := y
		if mirrorY {
			srcY = bounds.Max.Y - 1 - (y - bounds.Min.Y)
		}
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			srcX := x
			if mirrorX {
				srcX = bounds.Max.X - 1 - (x - bounds.Min.X)
			}
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}

func mirrorDirection(mirrorX, mirrorY bool) string {
	switch {
	case mirrorX && mirrorY:
		return "left-right-and-top-bottom"
	case mirrorX:
		return "left-right"
	default:
		return "top-bottom"
	}
}
