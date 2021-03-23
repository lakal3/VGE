//

package vmodel

import _ "embed"

//go:generate glslangValidator -V genmip.comp.glsl -o genmip.comp.spv

//go:embed genmip.comp.spv
var genmip_comp_spv []byte

//go:embed white.bin
var white_bin []byte
