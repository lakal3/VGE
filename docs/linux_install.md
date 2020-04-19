# Install VGE in Linux (experimental)

*Linux support is still experimetal and will be improved*

*VGE linux installation has been tested only on Ubuntu 18.04*

## Intall Vulkan SDK

Install Vulkan SDK an validate that you can run vbcube example in SDK. 

Getting working Vulkan driver on Linux machine seems bit more involved process that installing on in Windows.

This step is optional but recommended.

## libVGELib.so

Place prebuilt libVGELib.so so that Linux loader can find it. 
Alternatively [build](build_vgelib.md) libVGELib.so yourself.

Please follow you distributions specifics instructions on how to install shared libraries.

### Wayland (display server protocol)

Wayland is not yet supported, only X11 works for now!

## Testing

Try to run hello.go `go run hello.go` in examples/basic directory. If that works, try logo.go. 
If it fails (you don't see logo), you may have same bug that I encountered. 
Driver can only draw very small indirect mesh (witch makes kind of impossible to run more complex demos)

Laptops you mostly have to GPU:s, one integrated and one dedicated. Most samples support -dev switch to pic best suited one. 

