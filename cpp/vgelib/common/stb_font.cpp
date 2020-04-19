#include "vgelib/vgelib.hpp"

#define STB_TRUETYPE_IMPLEMENTATION
#include "third_party/stb/stb_truetype.h"


class StbFontLoader : public vge::FontLoader {
	friend struct FontLoader;
public:
	StbFontLoader(const uint8_t* font, size_t font_len) {
		stbtt_InitFont(&info, font, 0);
		scale = stbtt_ScaleForPixelHeight(&info, 48.0);
	};

	virtual void GetInfo(uint32_t codepoint, vge::CharInfo* charInfo) override {
		charInfo->extra = stbtt_FindGlyphIndex(&info, codepoint);
		if (charInfo->extra == 0) {
			return;
		}
		int x0, x1, y0, y1;
		stbtt_GetGlyphBitmapBox(&info, static_cast<int>(charInfo->extra), scale, scale, &x0, &y0, &x1, &y1);
		charInfo->height = y1 - y0;
		charInfo->width = x1 - x0;
		charInfo->offsetx = x0;
		charInfo->offsety = y0;
	}

	virtual void GetBitmap(vge::CharInfo* charInfo, uint8_t* bmp, size_t bmp_len) override {
		stbtt_MakeGlyphBitmap(&info, bmp, charInfo->width, charInfo->height, charInfo->width, scale, scale, static_cast<int>(charInfo->extra));
	}

	virtual void Dispose() override {
		delete this;
	}

	stbtt_fontinfo info;
	float scale;
};

vge::FontLoader *vge::FontLoader::NewFontLoader(const uint8_t* font, size_t font_len) 
{
	return new StbFontLoader(font, font_len);
}