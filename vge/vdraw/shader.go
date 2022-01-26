package vdraw

//go:generate glslangValidator -V draw.geom.glsl -o draw.geom.spv
//go:generate glslangValidator -V draw.vert.glsl -o draw.vert.spv
//go:generate glslangValidator -V draw.frag.glsl -o draw.frag.spv
//go:generate glslangValidator -DGLYPH=1 -V draw.frag.glsl -o draw_glyph.frag.spv
//go:generate glslangValidator -V copy.comp.glsl -o copy.comp.spv
//go:generate glslangValidator -V path.comp.glsl -o path.comp.spv

import (
	_ "embed"
)

//go:embed draw.geom.spv
var draw_geom_spv []byte

//go:embed draw.vert.spv
var draw_vert_spv []byte

//go:embed draw.frag.spv
var draw_frag_spv []byte

//go:embed draw_glyph.frag.spv
var draw_glyph_frag_spv []byte

//go:embed copy.comp.spv
var copy_comp_spv []byte

//go:embed path.comp.spv
var path_comp_spv []byte
