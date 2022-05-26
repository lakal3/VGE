//

package vapp

import _ "embed"

//go:generate glslangValidator -V tri.vert.glsl -o tri.vert.spv
//go:generate glslangValidator -V copy.frag.glsl -o copy.frag.spv

//go:embed copy.frag.spv
var copy_frag_spv []byte

//go:embed tri.vert.spv
var tri_vert_spv []byte
