//
package debugmat

import _ "embed"

//go:embed debugmat.frag.spv
var debugmat_frag_spv []byte

//go:embed debugmat.vert.spv
var debugmat_vert_spv []byte

//go:embed debugmat.vert_skinned.spv
var debugmat_vert_skinned_spv []byte
