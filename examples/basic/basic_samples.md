
# Basic Examples using vapp module 

Basic contains some basic samples on how to use VGE with highest level of abstraction. 

These example will use vapp package to setup for example Vulkan application, instance and device so you don't
have to take care of myriad of details you normally have to setup when creating a Vulkan application.

To run examples, just use go run. Just ensure that you have either:
- Built and installed VGELib.dll (.so) package 
- Copied this dll from {rootdir}/prebuild/{win/linux}/VGELib.dll

## Example in this directory

- hello.go - Open window and show very simple UI
- logo.go - In addition to hello, load logo from gltf 2.0 file
- logo_an.go - In addition to logo, adds some simple animation to scene

## Troubleshooting

Check that VGELib.dll (.so) is accessible from path.

Check that GOARCH is amd64. This is only supported architecture.

Install Vulkan SDK (TODO: add instructions) and see if you see any validation errors.
  