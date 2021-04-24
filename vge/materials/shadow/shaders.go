//

package shadow

import _ "embed"

//go:generate glslangValidator -V shadow.vert.glsl -o shadow.vert.spv
//go:generate glslangValidator -V -DSKINNED=1 shadow.vert.glsl -o shadow.vert_skin.spv
//go:generate glslangValidator -V shadow.geom.glsl -o shadow.geom.spv
//go:generate glslangValidator -V shadow.frag.glsl -o shadow.frag.spv
//go:generate glslangValidator -V point_shadow.vert.glsl -o point_shadow.vert.spv
//go:generate glslangValidator -V -DSKINNED=1 point_shadow.vert.glsl -o point_shadow.vert_skin.spv
//go:generate glslangValidator -V dir_shadow.vert.glsl -o dir_shadow.vert.spv
//go:generate glslangValidator -V -DSKINNED=1 dir_shadow.vert.glsl -o dir_shadow.vert_skin.spv
//go:generate glslangValidator -V point_shadow.frag.glsl -o point_shadow.frag.spv

//go:embed point_shadow.frag.spv
var point_shadow_frag_spv []byte

//go:embed point_shadow.vert.spv
var point_shadow_vert_spv []byte

//go:embed point_shadow.vert_skin.spv
var point_shadow_vert_skin_spv []byte

//go:embed dir_shadow.vert.spv
var dir_shadow_vert_spv []byte

//go:embed dir_shadow.vert_skin.spv
var dir_shadow_vert_skin_spv []byte

//go:embed shadow.frag.spv
var shadow_frag_spv []byte

//go:embed shadow.vert.spv
var shadow_vert_spv []byte

//go:embed shadow.vert_skin.spv
var shadow_vert_skin_spv []byte

//go:embed shadow.geom.spv
var shadow_geom_spv []byte
