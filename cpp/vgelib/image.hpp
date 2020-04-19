#pragma once

namespace vge {
	class ImageLoader: public Disposable {
		friend struct Static;
	public:
		ImageLoader();
		void Supported(const char* kind, size_t kind_len, bool& read, bool &write);
		void Load(const char* kind, size_t kind_len, void* bytes, size_t byte_len, Buffer *buffer);
		void Describe(const char* kind, size_t kind_len, ImageDescription* desc, void* bytes, size_t byte_len);
		void Save(const char* kind, size_t kind_len, ImageDescription* desc, Buffer* buffer, void* bytes, size_t byte_len, size_t& needLength);
	protected:
		virtual void Dispose() override;
	};
}