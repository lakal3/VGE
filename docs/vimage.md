# Image handling in VGE

The VGE implements a pluggable architecture for different image formats.
You can use vasset.RegisterImageLoader to register a new image format or replace an existing one.

## Supported image format and loaders

Currently supported image formats and loaders

### DDS

Microsoft DDS image format that support mipmaps and multiple image layers. DDS images supports a very wide range of different image formats.

DDS loader is implemented in native Go and is always registered.
DDS loader is very fast in processing images because the images are not packed (unless the image is in BC1-7 packed formats).
Images are also more or less like the GPU expects them to be.

There is also an excellent image converter Texconv [https://github.com/Microsoft/DirectXTex/wiki/Texconv]
that you can use to convert images from other formats to DDS and back.

Currently, the only way to use compressed images with VGE in GPU is to pack them with Texconv or some other tool into packed DDS formats.

_You can use Visual Studio 2019 or later to view DDS files._

**The DDS Loader can only read newer DDS files containing a DX10 header**. Texconv has a switch to force it on.

### PNG (Portable Network Graphics)

PNG images have two loaders available: PngLoader and NativeLoader.

PngLoader is using Go's image and png packages. Register this loader using pngloader.RegisterPngLoader.
Go's implementation of PNG loader can both read and write png images, but reading might be a bit slower than when using NativeLoader.

NativeLoader is implemented in C++ vgelib and uses internally stb_image.h image loaders.
NativeLoader is registered with vasset.RegisterNativeImageLoader. NativeLoader can only read PNG images.

You can use PngLoader to write images and NativeLoader to load images if you register PngLoader before.

### JPEG (Joint Photographic Experts Group)

JPEG images can be loaded with NativeLoaders. NativeLoaders are much faster in loading JPEG images than Go's own libraries.

There is currently no loader that can save JPEG images.

### HDR (High Dynamic Range)

HDR images in equirectangular projection are mostly used in the background. The NativeLoader supports loading HDR images.

Alternately, you can convert HDR images to (compressed) DDS using Texconv.


