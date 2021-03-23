//

package std

import _ "embed"

//go:generate glslangValidator -V std.vert.glsl -o std.vert.spv
//go:generate glslangValidator -V -DSKINNED=1 std.vert.glsl -o std.vert_skin.spv
//go:generate glslangValidator -V std.frag.glsl -o std.frag.spv

//go:embed std.frag.spv
var std_frag_spv []byte

//go:embed std.vert.spv
var std_vert_spv []byte

//go:embed std.vert_skin.spv
var std_vert_skin_spv []byte
