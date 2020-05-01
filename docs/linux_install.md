# Install VGE in Linux (experimental)

*Linux support is still experimental and will be improved*

*the VGE Linux installation has been tested only on Ubuntu 18.04*

## Install Vulkan SDK

Install Vulkan SDK and validate that you can run the vbcube example in SDK. 

Getting a working Vulkan driver on Linux machine seems bit more involved process than with Windows.

This step is optional but recommended.

## libVGELib.so

Place prebuilt libVGELib.so so that Linux loader can find it. 
Alternatively [build](build_vgelib.md) libVGELib.so yourself.

Please follow your distribution specific instructions on how to install shared libraries.

### Wayland (display server protocol)

Wayland is not yet supported, only X11 works for now!

## Testing

Try to run hello.go `go run hello.go` in examples/basic directory. If that works, try logo.go. 
If it fails (you don't see logo), you may have the same bug that I encountered. 
The driver can only draw a very small indirect mesh (which makes it kind of impossible to run more complex demos)

Laptops you mostly have two GPU:s, one integrated and one dedicated. Most samples support -dev switch to pic the best suited one. 

