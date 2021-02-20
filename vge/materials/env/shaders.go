//

package env

import _ "embed"

//go:generate glslangValidator.exe -V eqrect.vert.glsl -o eqrect.vert.spv
//go:generate glslangValidator.exe -V eqrect.frag.glsl -o eqrect.frag.spv
//go:generate glslangValidator.exe -V graybg.frag.glsl -o graybg.frag.spv
//go:generate glslangValidator.exe -V probe.comp.glsl -o probe.comp.spv
//go:generate glslangValidator.exe -V sph.comp.glsl -o sph.comp.spv

//go:embed eqrect.vert.spv
var eqrect_vert_spv []byte

//go:embed eqrect.frag.spv
var eqrect_frag_spv []byte

//go:embed graybg.frag.spv
var graybg_frag_spv []byte

//go:embed probe.comp.spv
var probe_comp_spv []byte

//go:embed sph.comp.spv
var sph_comp_spv []byte
