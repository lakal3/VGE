//
package debugmat

import _ "embed"

//go:generate glslangValidator -V debugmat.vert.glsl -o debugmat.vert.spv
//go:generate glslangValidator -V -DSKINNED=1 debugmat.vert.glsl -o debugmat.vert_skinned.spv
//go:generate glslangValidator -V debugmat.frag.glsl -o debugmat.frag.spv

//go:embed debugmat.frag.spv
var debugmat_frag_spv []byte

//go:embed debugmat.vert.spv
var debugmat_vert_spv []byte

//go:embed debugmat.vert_skinned.spv
var debugmat_vert_skinned_spv []byte
