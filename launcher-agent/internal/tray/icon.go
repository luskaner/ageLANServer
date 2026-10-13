package tray

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"sync"
)

// The icon is drawn rather than embedded.
//
// A tray icon is three shapes, and an embedded one is a binary file that has to
// be sourced, licensed, regenerated whenever the palette moves and, on the
// platforms that keep icons in resources, compiled into the executable. Drawing
// it costs a few hundred bytes of Go and leaves the icon in step with the code by
// construction.
//
// It is a globe because that is what the agent is about: a machine that was
// alone on a network is not any more. It is a solid disk rather than an outline
// because at the sixteen pixels the notification area actually uses, an outline
// is a ring of grey specks and a solid shape is still a shape.

// supersample is the factor each side is drawn at before being filtered down.
//
// Tray icons end up at sixteen physical pixels, and a shape drawn straight to
// that size has a staircase along its edge that reads as a mistake. Drawing big
// and averaging is what buys the soft edge the rest of the notification area
// has.
const supersample = 4

const (
	// diskRadius is the globe's radius as a fraction of the icon's side. The
	// margin is not decoration: the shell draws its own padding around the icon
	// and an edge flush with the side touches it.
	diskRadius = 0.46

	// equator is the half height of the horizontal stripe, as a fraction of the
	// radius.
	equator = 0.075

	// meridian is the half width of the vertical outline, as a fraction of the
	// distance to the curve itself rather than to a pixel, so that it stays the
	// same thickness whether it is crossing the middle of the disk or its edge.
	meridian = 0.06

	// meridianFlatten is how much narrower the meridian ellipse is than the disk.
	// A meridian as wide as the disk is a line, and the globe loses the curve
	// that makes it read as a sphere.
	meridianFlatten = 0.45
)

var (
	diskColour = color.RGBA{R: 0x2e, G: 0x9e, B: 0x5b, A: 0xff}
	lineColour = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
)

// iconSize is the size the icon is rendered at. The platform scales it to
// whatever the notification area asks for, and a source larger than that keeps
// the downscale honest on a high density display.
const iconSize = 32

// iconPNG is the rendered icon, encoded once.
var iconPNG = sync.OnceValue(func() []byte {
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, renderIcon(iconSize)); err != nil {
		// The only way png.Encode fails on a buffer that is not a file is a
		// failure to encode an image this package just built. There is nothing to
		// fall back to and nowhere to report it from, so an empty icon is the
		// honest answer: the tray comes up with no picture rather than not at all.
		return nil
	}
	return buffer.Bytes()
})

// renderIcon draws the globe and filters it down to size.
func renderIcon(size int) *image.RGBA {
	big := size * supersample
	centre := float64(big) / 2
	radius := diskRadius * float64(big)

	// Accumulated at full width and divided at the end: sixteen samples can add
	// up to more than a byte, so they cannot live in the image until averaged.
	pixels := make([]uint32, size*size*4)
	for y := 0; y < big; y++ {
		for x := 0; x < big; x++ {
			shade, ok := shadeAt(float64(x)+0.5-centre, float64(y)+0.5-centre, radius)
			if !ok {
				continue
			}
			offset := (y/supersample*size + x/supersample) * 4
			pixels[offset] += uint32(shade.R)
			pixels[offset+1] += uint32(shade.G)
			pixels[offset+2] += uint32(shade.B)
			pixels[offset+3] += uint32(shade.A)
		}
	}

	samples := uint32(supersample * supersample)
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for i := 0; i < size*size; i++ {
		offset := i * 4
		// Rounded rather than truncated: truncating loses the half of every
		// sample that the average lands on, which is a fifth of a bit on a fully
		// covered pixel and rather more on the edge that is the whole point.
		img.Pix[offset] = uint8((pixels[offset] + samples/2) / samples)
		img.Pix[offset+1] = uint8((pixels[offset+1] + samples/2) / samples)
		img.Pix[offset+2] = uint8((pixels[offset+2] + samples/2) / samples)
		img.Pix[offset+3] = uint8((pixels[offset+3] + samples/2) / samples)
	}
	return img
}

// shadeAt is the colour of one sample, and whether the sample is on the icon at
// all. The origin is the centre of the icon.
func shadeAt(dx float64, dy float64, radius float64) (color.RGBA, bool) {
	if dx*dx+dy*dy > radius*radius {
		return color.RGBA{}, false
	}
	if math.Abs(dy) <= equator*radius {
		return lineColour, true
	}
	// The meridian is the outline of an ellipse, so it is drawn as the distance
	// to that ellipse rather than as a stroke of a fixed width, which would come
	// out thick where the curve runs along the rows and thin where it crosses
	// them.
	along := math.Hypot(dx/(meridianFlatten*radius), dy/radius)
	if math.Abs(along-1) < meridian {
		return lineColour, true
	}
	return diskColour, true
}
