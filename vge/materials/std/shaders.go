//

package std

import _ "embed"

//go:generate glslangValidator -V std.vert.glsl -o std.vert.spv
//go:generate glslangValidator -V -DSKINNED=1 std.vert.glsl -o std.vert_skin.spv
//go:generate glslangValidator -V std.frag.glsl -o std.frag.spv
//go:generate glslangValidator -V defmat.vert.glsl -o defmat.vert.spv
//go:generate glslangValidator -V -DSKINNED=1 defmat.vert.glsl -o defmat.vert_skin.spv
//go:generate glslangValidator -V defmat.frag.glsl -o defmat.frag.spv

//go:embed std.frag.spv
var std_frag_spv []byte

//go:embed std.vert.spv
var std_vert_spv []byte

//go:embed std.vert_skin.spv
var std_vert_skin_spv []byte

//go:embed defmat.vert.spv
var defmat_vert_spv []byte

//go:embed defmat.frag.spv
var defmat_frag_spv []byte

//go:embed defmat.vert_skin.spv
var defmat_vert_skin_spv []byte
