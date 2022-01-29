package main

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/lakal3/vge/vge/vimgui"
	"strings"
)

const lipsum = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam ut mi id lorem tempus consequat vel mollis massa. Vestibulum in elit dui.
Quisque mattis consectetur dolor ut dapibus. Nunc fermentum metus quis augue congue, in tincidunt arcu egestas. Donec id laoreet magna.
Aliquam nisl nisl, tincidunt nec dignissim eu, condimentum vel tortor. Proin luctus nulla nisl, sit amet aliquet mauris malesuada a. 
Vestibulum arcu risus, viverra accumsan diam a, lobortis faucibus nisi. Cras efficitur, magna sit amet mollis finibus, 
nulla risus mattis massa, at hendrerit nisl nisl eget elit. Maecenas at elit at lacus suscipit lobortis. Morbi ut volutpat mauris.
Nunc et orci lobortis orci mattis tristique sed in est. Vivamus nec pellentesque velit.

Sed ut sapien mi. Nullam euismod quam nulla, vel posuere urna dapibus nec. Duis eleifend neque vel diam elementum tempus. Sed libero neque, 
efficitur sit amet quam id, porta iaculis nisi. Cras aliquam sed neque quis viverra. Nam pretium gravida sagittis. 
Sed id tortor in nulla malesuada vestibulum vitae eget neque. Nunc at massa ut lacus elementum facilisis.

Ut maximus accumsan urna non viverra. Nulla tincidunt gravida ligula, at tempus nisl mollis in. Sed ac imperdiet libero. 
Suspendisse fringilla odio mauris, id tincidunt odio condimentum id. In maximus porttitor congue. 
Aliquam a mi vitae mauris malesuada vehicula ac vitae mauris. Nunc eu elit in ipsum maximus auctor.

Mauris pharetra ut ligula sit amet accumsan. Integer mollis risus purus, vitae tincidunt purus aliquam at. 
Cras porta arcu a lorem pharetra, pulvinar vestibulum turpis lacinia. Sed odio elit, bibendum ut lacinia non, vulputate ut mauris.
Mauris at mauris vel lectus sagittis tincidunt. Aliquam erat volutpat. Nulla mollis libero nec augue cursus mattis.

Phasellus vel orci elit. Sed venenatis orci laoreet dignissim pretium. Proin ac magna diam. Duis feugiat, nibh sit amet vulputate porttitor, 
risus nunc porta dui, sed feugiat nunc sapien at nisl. Quisque porttitor, ligula suscipit convallis feugiat, ex urna mattis ante, 
in maximus erat magna vel dolor. Vivamus mollis ipsum sit amet bibendum lobortis. Class aptent taciti sociosqu ad litora torquent per
conubia nostra, per inceptos himenaeos. Praesent porta, orci eget faucibus bibendum, quam felis ullamcorper nisl, 
id fermentum lacus metus et nisl. Proin hendrerit porttitor ipsum, ut condimentum augue. In tincidunt erat vitae mi volutpat pellentesque. 
Mauris congue, lorem posuere tempus tincidunt, augue nibh convallis lectus, sed finibus lectus metus sed dolor. Quisque et justo erat. 
Mauris metus nisi, lobortis id ipsum at, mattis viverra quam. Morbi lectus nunc, mollis vel imperdiet id, convallis eu neque.`

var page2Offset mgl32.Vec2

func page2(fr *vimgui.UIFrame) {
	lines := strings.Split(lipsum+lipsum+lipsum+lipsum, "\n")
	s := mgl32.Vec2{1200, float32(len(lines) * 18)}
	vimgui.ScrollArea(fr, s, &page2Offset, func(uf *vimgui.UIFrame) {
		for _, l := range lines {
			fr.NewLine(-100, 18, 0)
			vimgui.Label(uf, l)
		}
	})
}
