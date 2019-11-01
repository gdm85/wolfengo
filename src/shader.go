/*
WolfenGo - https://github.com/gdm85/wolfengo
Copyright (C) 2016~2019 gdm85

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
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-gl/gl/v2.1/gl"
)

type Shader struct {
	program  uint32
	uniforms map[string]int32
}

func NewShader(withUpdateUniforms bool) (*Shader, error) {
	s := &Shader{}
	s.program = gl.CreateProgram()
	if s.program == 0 {
		return nil, errors.New("shader creation failed: could not find valid memory location when creating program")
	}

	if withUpdateUniforms {
		s.uniforms = make(map[string]int32, 0)
	}

	return s, nil
}

func (s *Shader) bind() {
	gl.UseProgram(s.program)
}

func (s *Shader) addUniform(uniformName string) error {
	csource, free := gl.Strs(uniformName + "\000")
	defer free()
	uniformLocation := gl.GetUniformLocation(s.program, *csource)

	if uniformLocation == -1 {
		return fmt.Errorf("could not find uniform: %s", uniformName)
	}

	s.uniforms[uniformName] = uniformLocation
	return nil
}

func (s *Shader) getProgramInfoLog(context string) error {
	var logLength int32
	gl.GetProgramiv(s.program, gl.INFO_LOG_LENGTH, &logLength)

	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetProgramInfoLog(s.program, logLength, nil, gl.Str(log))

	return fmt.Errorf("%s: %s", context, log)
}

func (s *Shader) getShaderInfoLog(shader uint32, context string) error {
	var logLength int32
	gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

	return fmt.Errorf("%s: %s", context, log)
}

func (s *Shader) compile() error {
	gl.LinkProgram(s.program)
	var result int32
	gl.GetProgramiv(s.program, gl.LINK_STATUS, &result)
	if result == gl.FALSE {
		return s.getProgramInfoLog("shader linking error")
	}
	gl.ValidateProgram(s.program)
	gl.GetProgramiv(s.program, gl.VALIDATE_STATUS, &result)
	if result == gl.FALSE {
		return s.getProgramInfoLog("shader validation error")
	}

	return nil
}

func (s *Shader) addProgram(text string, typ uint32) error {
	shader := gl.CreateShader(typ)
	if shader == 0 {
		return errors.New("could not find valid memory location when adding shader")
	}
	cStr, free := gl.Strs(text + "\000")
	defer free()
	gl.ShaderSource(shader, 1, cStr, nil)
	gl.CompileShader(shader)

	var result int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &result)
	if result == gl.FALSE {
		return s.getShaderInfoLog(shader, "shader compilation error")
	}

	gl.AttachShader(s.program, shader)
	return nil
}

func (s *Shader) addProgramFromFile(fileName string, typ uint32) error {
	data, err := ioutil.ReadFile("./res/shaders/" + fileName)
	if err != nil {
		return err
	}

	return s.addProgram(string(data), typ)
}

func (s *Shader) setUniform(uniformName string, value Vector3f) {
	gl.Uniform3f(s.uniforms[uniformName], value.X, value.Y, value.Z)
}

func (s *Shader) setUniformM(uniformName string, value Matrix4f) {
	floats := value.asArray()

	gl.UniformMatrix4fv(s.uniforms[uniformName], 1, true, &floats[0])
}

func (s *Shader) updateUniforms(projectedMatrix Matrix4f, material *Material) {
	if s.uniforms == nil {
		return
	}

	if material.texture != nil {
		material.texture.bind()
	} else {
		gl.BindTexture(gl.TEXTURE_2D, 0)
	}

	s.setUniformM("transform", projectedMatrix)
	s.setUniform("color", material.color)
}
