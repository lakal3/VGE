//

package emissive

import _ "embed"

//go:generate glslangValidator -V emissive.vert.glsl -o emissive.vert.spv
//go:generate glslangValidator -V -DSKINNED=1 emissive.vert.glsl -o emissive.vert_skin.spv
//go:generate glslangValidator -V emissive.frag.glsl -o emissive.frag.spv

//go:embed emissive.frag.spv
var emissive_frag_spv []byte

//go:embed emissive.vert.spv
var emissive_vert_spv []byte

//go:embed emissive.vert_skin.spv
var emissive_vert_skin_spv []byte
