# VGE User interface

VGE support UI suitable for games and simulations. It still lack some features like clipboard control typically found in business applications.

Nearly all examples set up some kind of UI.

## UIView

UIView is root control of UI. To show an UI, you must bring UIView to scene.

Actual UI is constructed with UI controls. UI controls in VGE follow same philosophy used in some modern UI frameworks: 
Each individual controls are very simple and only do one task. In Go there is no inheritance which makes building 10 level deep inherited controls quite challenging anyway (and mostly quite useless).
You can then combine simple UI controls to build smarter UIs.
 
For example there is no margin or padding settings in normal controls. However there is generic Padding control
that can inset it's content based on requested padding. 

## Layout

VGE controls supports simple measurement system. Each component can calculate optimal size with given UI width.
Some control like Stacks use this information to position controls. 

If you need more complex layout Canvas control supports scaling and anchoring content based on design and actual size. 
You can of cause embed canvas inside canvas to get more complex layouts. 
If that is not enough it is most likely quite easy to make custom container that can handle required layout calculations.

## Controls

VGE support basic UI controls including. UI example project show most of controls in VGE.

### Basic controls

#### Label

Control that draw Text.

#### TextBox

User editable single line text area. 
Supports change event when user changes text box content.
 
#### Button and MenuButton

Button control has content (typically a label but can be anything) and Click action that will be activated if user clicks button. 
MenuButton is like Button but typically has different representation (Style).

#### Padding

Pads inner content

#### Conditional

Allows hiding content with a boolean flag. VGE controls are this simple, there is not event visibility flag on normal controls. 
You use Conditional control to implement that. 

#### ToggleButton

ToggleButton toggles inner content when user clicks control. 
ToggleButton also support OnChanged event to detect state change. ToggleButton is typically used to draw radio and check boxes.

#### HSlider and VSlider

Horizontan and vertical slider that support changed event when user drags slider.

#### Panel

Panel that contains child controls. Typically to level element of UIView.

#### Sizer

Allow setting minimum and maximum size for content. 

_Some parent controls may disregard size request and draw control with different size_


### Layout controls

Layout controls can have multiple children. 

#### HStack and VStack

Layouts children horizontally or vertically. Stack controls use child size measurement when
arranging child controls. Both stack support padding between elements. 

#### Canvas

More complex layout control supporting anchoring and resizing child controls based on design size vs actual size.

Each child have a field that tells:
 - How must of canvas size change should be applied to children position.
 - How must of canvas size change should be applied to children size. 
 
#### ScrollViewer

"Infinite" area that supports scrolling content if content don't fit into viewable area.
ScrollArea only support single child that is typically some content control.


## Themes

VGE Controls them shelf don't contain any representation! Instead you must assign a Theme for each UIView. 
VGE controls will use given theme to request how they should draw them selves. 
This is bit similar idea that is implemented at least in WPF (Windows Presentation Foundation) although this implementation is much more simpler.

Control will as theme to provide a Style for control. Style is then used to measure and draw control. 
Styles typically use GlyphSets to draw actual UI elements and fonts.

VGE contains one standard theme (mintheme) that only used vector graphics to draw controls. 
Mintheme allows some customization like changing default font. 

UI example project has also sample custom theme theme3D. Theme3D is an example how to build your own theme using single color bitmap glyphs.  
   
## Glyphs and GlyphSets

Glyphs are implemented in vglyph model. Idea is to pack multiple small images into GlyphSet. Then VGE can draw individual glyphs from glyph sets.

GlyphSets are mainly either shapes for UI controls or fonts. VGE draws user interfaces by drawing glyphs.

GlyphSet can have one of following formats:
- Signed depth field. Signed depth fields are built from vectors using vector builder. Vectors can be lines or quadratic bezier curves. Depth fields have only on / off settings.
Signed depth field allow more smooth scaling and is optimal for TTF fonts 
- Single color + alpha. 
- Full color (rgba8)

Glyphs can also have border, either top/botton, left/right, both or neither. 
Font typically don't have border areas. Most ui glyphs like button will have both border areas.

![Glyph areas](glyph_split.png)

When glyph is draw, we will give glyph location and color. Additionally we will give size of borders for glyphs that has border area. 
Resizing bordered glyph but keeping border size constant will stretch image so that borders will keep they thickness.

When drawing a glyph, colors are handle depending on GlyphSet format
- For signed depth field, pixels have set (<0.5) for fore color and unset (0>0.5) back color. 
  Glyph shader will smoothstep between on/off values to make shaper more sharp looking,  
- For single color + alpha, color intensity controls ratio between fore color (1) and back color. Alpha channel is multiplied with images alpha color.
- For rgba glyphs, color settings are ignored. Colors are fetched directly from image.

So full RGBA glyph has least rendering options but can show colorful glyphs. Signed depth fields scales best but support only on/off color. 
Single color glyph is something in between.

## Fonts

VectorBuilder can directly build glyph set from TTF (TrueType) font file. 
See for example glTFviewer for example on how to load a font file into GlyphSet.



