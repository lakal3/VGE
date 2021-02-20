//

package predepth

import _ "embed"

//go:generate glslangValidator.exe -V predepth.vert.glsl -o predepth.vert.spv
//go:generate glslangValidator.exe -V -DSKINNED=1 predepth.vert.glsl -o predepth.vert_skin.spv
//go:generate glslangValidator.exe -V predepth.frag.glsl -o predepth.frag.spv

//go:embed predepth.frag.spv
var predepth_frag_spv []byte

//go:embed predepth.vert.spv
var predepth_vert_spv []byte

//go:embed predepth.vert_skin.spv
var predepth_vert_skin_spv []byte
