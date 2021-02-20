//

package vglyph

import _ "embed"

//go:generate glslangValidator.exe -V copy.comp.glsl -o copy.comp.spv
//go:generate glslangValidator.exe -V copy_rgb.comp.glsl -o copy_rgb.comp.spv
//go:generate glslangValidator.exe -V vdepth.comp.glsl -o vdepth.comp.spv
//go:generate glslangValidator.exe -V glyph.vert.glsl -o glyph.vert.spv
//go:generate glslangValidator.exe -V glyph.frag.glsl -o glyph.frag.spv

//go:embed copy.comp.spv
var copy_comp_spv []byte

//go:embed copy_rgb.comp.spv
var copy_rgb_comp_spv []byte

//go:embed vdepth.comp.spv
var vdepth_comp_spv []byte

//go:embed glyph.frag.spv
var glyph_frag_spv []byte

//go:embed glyph.vert.spv
var glyph_vert_spv []byte
