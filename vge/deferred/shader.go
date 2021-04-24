package deferred

import _ "embed"

//go:generate glslangValidator -V lights.vert.glsl -o lights.vert.spv
//go:generate glslangValidator -V lights.frag.glsl -o lights.frag.spv

//go:embed lights.vert.spv
var lights_vert_spv []byte

//go:embed lights.frag.spv
var lights_frag_spv []byte
