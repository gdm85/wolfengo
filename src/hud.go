package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"time"

	xfont "golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/gdm85/wolfengo/src/gl"
)

const (
	hudLevelDisplayDuration = 3 * time.Second
	hudFontSizeLarge        = 72.0
	hudFontSizeSmall        = 40.0
	hudFontDPI              = 72.0
	hudPadding              = 8
	hudMargin               = float32(0.02) // NDC units from the screen edge
)

var wolfensteinHudColor = color.RGBA{R: 98, G: 19, B: 20, A: 255}

// hudText bundles a GL texture with the pixel dimensions of its source image.
type hudText struct {
	tex  *Texture
	imgW int
	imgH int
}

// HUD manages all screen-space overlays: health, level announcement, game over.
type HUD struct {
	quadMesh  Mesh
	whiteMat  *Material
	faceLarge xfont.Face
	faceSmall xfont.Face

	health    *hudText
	healthVal int

	levelAnnounce *hudText
	levelHideAt   time.Time

	gameOverText *hudText
	isGameOver   bool
}

// loadHUD creates a HUD for the given level number.
// Font is loaded from res/fonts/hud.ttf; falls back to the embedded Go Regular.
func loadHUD(levelNum uint) (*HUD, error) {
	faceLarge, err := loadHUDFace(hudFontSizeLarge)
	if err != nil {
		return nil, fmt.Errorf("HUD large font: %w", err)
	}
	faceSmall, err := loadHUDFace(hudFontSizeSmall)
	if err != nil {
		return nil, fmt.Errorf("HUD small font: %w", err)
	}

	h := &HUD{
		quadMesh:  buildHUDQuadMesh(),
		faceLarge: faceLarge,
		faceSmall: faceSmall,
	}
	whiteTex := NewWhiteTexture()
	h.whiteMat = NewMaterial(whiteTex)

	h.setHealth(defaultPlayer.maxHealth)
	h.showLevel(levelNum)

	return h, nil
}

func loadHUDFace(size float64) (xfont.Face, error) {
	fontData, err := ioutil.ReadFile("res/fonts/hud.ttf")
	if err != nil {
		fontData = goregular.TTF // embedded fallback
	}
	f, err := opentype.Parse(fontData)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(f, &opentype.FaceOptions{
		Size:    size,
		DPI:     hudFontDPI,
		Hinting: xfont.HintingFull,
	})
}

// ---- HUD state mutators -------------------------------------------------------

func (h *HUD) showLevel(n uint) {
	h.levelAnnounce = makeHUDText(fmt.Sprintf("Level %d", n), h.faceLarge)
	h.levelHideAt = time.Now().Add(hudLevelDisplayDuration)
}

func (h *HUD) showGameOver() {
	if !h.isGameOver {
		h.isGameOver = true
		h.gameOverText = makeHUDText("GAME OVER", h.faceLarge)
	}
}

func (h *HUD) setHealth(hp int) {
	if h.health != nil && h.healthVal == hp {
		return
	}
	h.health = makeHUDText(fmt.Sprintf("%d", hp), h.faceSmall)
	h.healthVal = hp
}

// ---- Rendering ---------------------------------------------------------------

func (h *HUD) render(shader *Shader, screenW, screenH float32) {
	gl.Disable(gl.DEPTH_TEST)
	defer gl.Enable(gl.DEPTH_TEST)

	// Health: bottom-left, anchored to the left and bottom edges.
	if h.health != nil {
		ndcW := float32(h.health.imgW) / screenW * 2
		ndcH := float32(h.health.imgH) / screenH * 2
		cx := -1 + ndcW/2 + hudMargin
		cy := -1 + ndcH/2 + hudMargin
		h.drawText(shader, h.health, cx, cy, screenW, screenH)
	}

	// Level announcement: top-center.
	if h.levelAnnounce != nil && !time.Now().After(h.levelHideAt) {
		ndcH := float32(h.levelAnnounce.imgH) / screenH * 2
		h.drawText(shader, h.levelAnnounce, 0, 1-ndcH/2-hudMargin, screenW, screenH)
	}

	// Game over: centered.
	if h.isGameOver && h.gameOverText != nil {
		h.drawText(shader, h.gameOverText, 0, 0, screenW, screenH)
	}
}

// drawText draws a hudText quad centered at NDC (cx, cy).
func (h *HUD) drawText(shader *Shader, ht *hudText, cx, cy, screenW, screenH float32) {
	ndcW := float32(ht.imgW) / screenW * 2
	ndcH := float32(ht.imgH) / screenH * 2

	var trans, scale Matrix4f
	trans.initTranslation(cx, cy, 0)
	scale.initScale(ndcW, ndcH, 1)
	m := trans.mul(scale)

	mat := NewMaterial(ht.tex)
	// color stays {1,1,1} from NewMaterial default — no tint on the text texture
	shader.updateUniforms(m, mat)
	h.quadMesh.draw()
}

// ---- Helpers -----------------------------------------------------------------

// buildHUDQuadMesh returns a unit quad in NDC [-0.5,0.5]² with UV (0,0) at the
// top-left vertex so text images appear right-side up (same convention as sprites).
func buildHUDQuadMesh() Mesh {
	vertices := []*Vertex{
		{Vector3f{-0.5, -0.5, 0}, Vector2f{0, 1}, Vector3f{}},
		{Vector3f{-0.5, +0.5, 0}, Vector2f{0, 0}, Vector3f{}},
		{Vector3f{+0.5, +0.5, 0}, Vector2f{1, 0}, Vector3f{}},
		{Vector3f{+0.5, -0.5, 0}, Vector2f{1, 1}, Vector3f{}},
	}
	indices := []int32{0, 1, 2, 0, 2, 3}
	return NewMesh(vertices, indices, false)
}

// makeHUDText renders text into an image and uploads it as a GL texture.
func makeHUDText(text string, face xfont.Face) *hudText {
	bounds, _ := xfont.BoundString(face, text)

	imgW := (bounds.Max.X - bounds.Min.X).Ceil() + hudPadding*2
	imgH := (bounds.Max.Y - bounds.Min.Y).Ceil() + hudPadding*2
	if imgW < 1 {
		imgW = 1
	}
	if imgH < 1 {
		imgH = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))
	draw.Draw(img, img.Bounds(), image.Transparent, image.Point{}, draw.Src)

	d := xfont.Drawer{
		Dst:  img,
		Src:  image.NewUniform(wolfensteinHudColor),
		Face: face,
		Dot: fixed.Point26_6{
			X: fixed.I(hudPadding) - bounds.Min.X,
			Y: fixed.I(hudPadding) - bounds.Min.Y,
		},
	}
	d.DrawString(text)

	return &hudText{
		tex:  newTextureFromRGBA(img),
		imgW: imgW,
		imgH: imgH,
	}
}

// newTextureFromRGBA uploads an image.RGBA to the GPU using the same row-order
// convention as loadTexture (first row = top of image → maps to GL texture s,t=*,0).
func newTextureFromRGBA(img *image.RGBA) *Texture {
	b := img.Bounds()
	w, h := int32(b.Dx()), int32(b.Dy())

	buf := make([]byte, w*h*4)
	for y := int32(0); y < h; y++ {
		for x := int32(0); x < w; x++ {
			c := img.RGBAAt(int(x), int(y))
			i := (y*w + x) * 4
			buf[i], buf[i+1], buf[i+2], buf[i+3] = c.R, c.G, c.B, c.A
		}
	}

	t := &Texture{}
	gl.GenTextures(1, &t.ID)
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, w, h, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(buf))
	return t
}
