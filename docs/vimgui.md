
# VIMGUI - immediate mode user interface

VIMGUI = VGE immediate mode graphical user interface

## Background

### Retained mode user interface

Typically, user interface libraries like Windows Forms, GTK, WPF (Windows presentation framework) etc..
uses hierarchy of UI controls (often called Widgets). Controls are stateful "objects", even
some implementation languages like are not object-oriented.
When ever we want to make a change we modify object hierarchy or object state.
These kinds of user interfaces are usually called retained more user interfaces.

### Immediate mode user interface

Immediate mode user interface uses quite different approach. 
Whenever UI is draw, immediate mode UI will call user defined function
that should draw user interface in current state. For example:

```go
var page int
// kPage is IDs for page controls
var kPage = vk.NewKeys(10)

func draw(fr *vimgui.UIFrame) {
	// Negative column width is width in percentage of available draw area. -95 = 95%
	fr.NewLine(-95, 22, 0)
	vimgui.TabButton(fr, kPage, "Base controls", 0, &page)
	fr.NewLine(-95, 22, 2)
	vimgui.TabButton(fr, kPage+1, "Scroll area", 1, &page)
	fr.NewLine(-95, 22, 3)
	vimgui.TabButton(fr, kPage+2, "Shapes", 2, &page)
	fr.NewLine(-95, 3, 3)
	switch page {
	case 0:
	    // Draw base controls page
    case 1:
		// Draw scroll area page
	...	
	} 
```

Draw function will draw UI. Depending on the value of page draw function will
conditionally draw different content.

Also, change of value like page will happend when control is draw. 
This means that most immediate mode active return a boolean value indicating if value was changed just now or if button was pressed.


### Why immediate mode UI?

VGE already has [VUI](vui.md) retained mode UI. VIMGUI will replace VUI.

Retained mode UI have several challenges in real time rendering:
- They are typically single threaded. All state changes must be done in specific thread or using some kind of locking.
  (current VUI in VGE use function that are called when we can safely change UI for state change)
- Modifying UI state is often more complex that just branching code
- Immediate mode control (widgets) are typically much simpler than retained mode controls (widgets).
 Often just few nested function calls.
- There is no complex event wiring at all. Most control will just return value to indicate some action.


### Similar UI frameworks

Although most of the common UI frameworks shipped with different operating systems are retained mode UI, immediate mode UI is not new idea.
Maybe the best known implementation of immediate mode UI is [Dear ImGui](https://github.com/ocornut/imgui).
Also, multi-platform UI framework [Flutter](https://flutter.dev/) uses immediate mode approach, although it uses objects and methods to draw UI.

VIMGUI has borrowed several ideas from those frameworks.

## Implementation

VIMGUI uses new vapp.ViewWindow to draw user interface. 

You must initialize new UI view using vimgui.NewView and attach view to Window.
Each View will take drawing function to draw actual content of UI view. 
Each view also need a theme. Theme is used to set visual appearance of controls.

You can have several UI or any other views attached to one ViewWindow.  

### Drawing

All standard controls use [vdraw](vdraw.md) vector drawing to actually draw control content.

### Control ID

All non-static controls in VIMGUI need control ID. Control ID is used to handle internal state for controls. 
Draw functions should ensure that same control will always use same ID. 

Although most of the controls don't currently need any state, state will be used when VGE styling will implement animation (coming later).
TextBox based controls will not work without proper state and proper control ID. 
They need to store caret location and current selection when control has focus.

*Use can use vk.NewHashKey if you like to use control names instead of raw keys*

### UIFrame

Each draw call will receive UI frame. This structure is used to control several aspects of drawing:
- UIFrame contains information about mouse position and mouse clicks. 
  Controls can check if mouse is inside control (hover) and if mouse was clicked or dragged.
- UIFrame contains last keyboard event and current focus since previous draw.
- UIFrame allows controls to have internal state (with Control ID)
- UIFrame contains current DrawArea. DrawArea is typically size of View but some content controls can adjust this area.
- UIFrame contains ControlArea. Each standard control uses ControlArea to position it in UI
- UIFrame contains tags. Each standard control add tags to list of tags used to search styles for controls. See styles and themes.

UIFrame have several methods to Push different values. 
Use Pop to restore values. Pop call are often good candidates for defer calls.


### Layout

All standard controls draw them at location specified by ControlArea in UIFrame. 
There are few helper methods to move control position to next column or next line. 
For more advances layouts, just use DrawArea and ControlArea to calculate position for control(s) 
before drawing them. In most cases this is much simpled that using complex UI controls.


## Styles and themes

### Styles

Each style is identified by a named struct like vimgui.BackgroundColor or vimgui.BorderColor. 
Standard controls will then use name of struct to access specific style. 
You can add new styles for your controls just by declaring a structure.

### Theme 

Theme is collection of priorities and tagged styles. When accessing a style from a theme, 
control will give style name (using default value if given structure type) and
with list of tags.

Tags are very similar to cascade style sheet classes in html. 

```go
	Theme.AddStyle(root)
	label := vimgui.Style{Tags: []string{"*label"}, Priority: 1}
	label.Set(vimgui.ForeColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.75, 0.75, 0.75, 1})})
	Theme.AddStyle(label)

	btn := vimgui.Style{Tags: []string{"*button"}, Priority: 1}
	btn.Set(vimgui.BorderRadius{Corners: vdraw.UniformCorners(8)}, vimgui.BorderThickness{Edges: vdraw.UniformEdge(2)})
	Theme.AddStyle(btn)
	Theme.Add(1, vimgui.Tags(":hover"),
		vimgui.BackgroudColor{Brush: vdraw.SolidColor(mgl32.Vec4{0.5, 0.5, 0.5, 0.2})})
    Theme.Add(20, vimgui.Tags("error"),
        vimgui.ForeColor{Brush: vdraw.SolidColor(mgl32.Vec4{1, 0.2, 0.2, 1})},
        vimgui.BorderColor{Brush: vdraw.SolidColor(mgl32.Vec4{1, 0.2, 0.2, 1})},
    )
    Theme.Add(10, vimgui.Tags("alert", "dialog", "*panel"),
        vimgui.BorderColor{Brush: vdraw.SolidColor(mgl32.Vec4{1, 0.2, 0.2, 1})})

```

For example, if we have following theme (from mintheme.go):
- if we have tag *label, 
 foreground color (ForeColor) will have value of light gray if we have at least tag "*label".
- Similarly we will have BorderRadius uniform 8 pixels for "*button".
- If we have tags "*label" and "error", ForeColor will be red because tag error has higher priority.
- Border color will be red if we have at least tags "alert", "dialog" and "*panel"

By convention, tags staring with * are reserved for named controls. 
Tags starting with : are UI states like :focus or :hover.

## Standard controls

See example gallery for standard controls and how to use them.

