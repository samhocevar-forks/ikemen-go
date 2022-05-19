package main

import (
	"runtime"

	gl "github.com/fyne-io/gl-js"
)

type Texture struct {
	handle gl.Texture
	size float64
}

// Generate a new texture name
func newTexture() (t *Texture) {
	t = &Texture{handle: gl.CreateTexture()}
	UpdateStat(StatTextureCount, 1)
	runtime.SetFinalizer(t, (*Texture).finalizer)
	return
}

// Bind a texture and upload texel data to it
func (t *Texture) SetData(width, height, depth int32, filter bool, data []byte) {
	interp := int(gl.NEAREST)
	if filter {
		interp = gl.LINEAR
	}

	size := float64(width * height) / (1024 * 1024)
	format := gl.Enum(gl.LUMINANCE)
	if depth == 24 {
		size *= 4
		format = gl.RGB
	} else if depth == 32 {
		size *= 4
		format = gl.RGBA
	}

	gl.BindTexture(gl.TEXTURE_2D, t.handle)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, int(width), int(height),
		format, gl.UNSIGNED_BYTE, data)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, interp)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, interp)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// Update texture stats
	UpdateStat(StatTextureMemory, size - t.size)
	t.size = size
}

func (t *Texture) finalizer() {
	UpdateStat(StatTextureCount, -1)
	UpdateStat(StatTextureMemory, -t.size)
	sys.mainThreadTask <- func() {
		gl.DeleteTexture(t.handle)
	}
}
