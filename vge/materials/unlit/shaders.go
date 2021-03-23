//

package unlit

import _ "embed"

//go:generate glslangValidator -V unlit.vert.glsl -o unlit.vert.spv
//go:generate glslangValidator -V -DSKINNED=1 unlit.vert.glsl -o unlit.vert_skin.spv
//go:generate glslangValidator -V unlit.frag.glsl -o unlit.frag.spv

//go:embed unlit.frag.spv
var unlit_frag_spv []byte

//go:embed unlit.vert.spv
var unlit_vert_spv []byte

//go:embed unlit.vert_skin.spv
var unlit_vert_skin_spv []byte
