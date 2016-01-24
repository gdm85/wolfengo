/* WolfenGo - https://github.com/gdm85/wolfengo
Copyright (C) 2016 gdm85

This program is free software; you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation; either version 2 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License along
with this program; if not, write to the Free Software Foundation, Inc.,
51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.
*/
package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"

	"github.com/go-gl/gl/v2.1/gl"
)

type Texture struct {
	ID uint32
}

type textureError struct {
	fileName string
	err      error
}

func (te textureError) Error() string {
	return fmt.Sprintf("loadTexture(%s): %v", te.fileName, te.err)
}

func NewTexture(fileName string) (*Texture, error) {
	t := &Texture{}
	var err error
	t.ID, err = loadTexture(fileName)
	if err != nil {
		return nil, textureError{fileName, err}
	}
	return t, nil
}

func loadTexture(fileName string) (uint32, error) {
	imgFile, err := os.Open("./res/textures/" + fileName)
	if err != nil {
		return 0, err
	}

	imgCfg, _, err := image.DecodeConfig(imgFile)
	if err != nil {
		return 0, err
	}
	_, err = imgFile.Seek(0, 0)
	if err != nil {
		return 0, err
	}

	w, h := int32(imgCfg.Width), int32(imgCfg.Height)

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return 0, err
	}

	buffer := make([]byte, w*h*4)
	index := 0
	for y := 0; y < int(h); y++ {
		for x := 0; x < int(w); x++ {
			pixel := img.At(x, y).(color.NRGBA)
			buffer[index] = pixel.R
			buffer[index+1] = pixel.G
			buffer[index+2] = pixel.B
			buffer[index+3] = pixel.A

			index += 4
		}
	}

	var texture uint32

	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA8,
		w,
		h,
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(buffer))

	return texture, nil
}

func (t *Texture) bind() {
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
}
