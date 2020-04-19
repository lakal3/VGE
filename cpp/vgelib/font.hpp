#pragma once


namespace vge {
	struct FontLoader : public Disposable {
		friend struct Static;
		static FontLoader* NewFontLoader(const uint8_t* font, size_t font_len);
		virtual void GetInfo(uint32_t codepoint, CharInfo* info) = 0;
		virtual void GetBitmap(CharInfo* charInfo, uint8_t* bmp, size_t bmp_len) = 0;
	};
}