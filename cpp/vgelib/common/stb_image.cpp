#include "vgelib/vgelib.hpp"

#define STBI_WINDOWS_UTF8
#define STB_IMAGE_IMPLEMENTATION
#define STBI_ONLY_PNG
#define STBI_ONLY_JPEG
#define STBI_ONLY_HDR
#define STBI_NO_STDIO

#include "third_party/stb/stb_image.h"

vge::ImageLoader::ImageLoader() {
}

void vge::ImageLoader::Supported(const char* kind, size_t kind_len, bool& read, bool &write) {
	std::string sKind(kind, kind_len);
	write = false;
	if (sKind == "png" || sKind == "jpg" || sKind == "jpeg" || sKind == "hdr") {
		read = true;
	} else {
		read = false;
	}
}

void vge::ImageLoader::Load(const char* kind, size_t kind_len, void* bytes, size_t byte_len, Buffer* buffer) {
	int x, y, comp;
	std::string sKind(kind, kind_len);
	void* toPtr;
	void* fromPtr;
	buffer->GetPtr(toPtr);
	size_t len = 0;
	if (sKind == "hdr") {
		fromPtr = stbi_loadf_from_memory(static_cast<stbi_uc*>(bytes), static_cast<int>(byte_len), &x, &y, &comp, 4);
		len = static_cast<size_t>(x) * static_cast<size_t>(y) * 4 * 4;
	} else {
		fromPtr = stbi_load_from_memory(static_cast<stbi_uc*>(bytes), static_cast<int>(byte_len), &x, &y, &comp, 4);
		len = static_cast<size_t>(x) * static_cast<size_t>(y) * 4;
	}
	if (buffer->getSize() < len) {
		throw std::runtime_error("Buffer too short");
	}
	memcpy(toPtr, fromPtr, len);
	STBI_FREE(fromPtr);
}

void vge::ImageLoader::Describe(const char* kind, size_t kind_len, ImageDescription* desc, void* bytes, size_t byte_len) {
	int x, y, comp;
	stbi_info_from_memory(static_cast<stbi_uc*>(bytes), static_cast<int>(byte_len), &x, &y, &comp);
	desc->Depth = 1;
	desc->Width = x;
	desc->Height = y;
	desc->Layers = 1;
	desc->MipLevels = 1;
	desc->Format = vk::Format::eR8G8B8A8Unorm;
	std::string sKind(kind, kind_len);
	if (sKind == "hdr") {
		desc->Format = vk::Format::eR32G32B32A32Sfloat;
	}
}

struct SaveContext {
	std::vector<char> bytes;
	size_t regSize = 0;
	bool write = false;
};

void writeToContext(void* context, void* data, int size) {
	auto sc = reinterpret_cast<SaveContext*>(context);
	sc->regSize += size;
	if (sc->write) {
		auto cd = static_cast<char*>(data);
		sc->bytes.insert(sc->bytes.end(), cd, cd + size);
	}
}

void vge::ImageLoader::Save(const char* kind, size_t kind_len, ImageDescription* desc, Buffer* buffer, void* bytes, size_t byte_len, size_t& needLength) {
	throw std::runtime_error("Native loader don't implement save");

}

void vge::ImageLoader::Dispose()
{
	delete this;
}
