//

package shadow

import _ "embed"

//go:generate glslangValidator.exe -V shadow.vert.glsl -o shadow.vert.spv
//go:generate glslangValidator.exe -V -DSKINNED=1 shadow.vert.glsl -o shadow.vert_skin.spv
//go:generate glslangValidator.exe -V shadow.geom.glsl -o shadow.geom.spv
//go:generate glslangValidator.exe -V shadow.frag.glsl -o shadow.frag.spv

//go:embed shadow.frag.spv
var shadow_frag_spv []byte

//go:embed shadow.vert.spv
var shadow_vert_spv []byte

//go:embed shadow.vert_skin.spv
var shadow_vert_skin_spv []byte

//go:embed shadow.geom.spv
var shadow_geom_spv []byte
