//

package noise

import _ "embed"

//go:generate glslangValidator.exe -V fire.vert.glsl -o fire.vert.spv
//go:generate glslangValidator.exe -V fire.frag.glsl -o fire.frag.spv

//go:embed fire.frag.spv
var fire_frag_spv []byte

//go:embed fire.vert.spv
var fire_vert_spv []byte
