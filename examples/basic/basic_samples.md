
# Basic Examples using the vapp module

Basic_samples contains some basic samples on how to use the VGE with the highest level of abstraction.

These examples will use the vapp package to setup, for example, the Vulkan application, instance and device so you do not
have to take care of the many details you would normally have to setup when creating a Vulkan application.

To run the examples, just use go run. Ensure that you have either:
- Built and installed vgelib.dll (.so) package
- Copied this dll from {rootdir}/prebuild/{win/linux}/vgelib.dll

## Examples in this directory

- hello.go - Opens a window and shows a very simple UI
- logo.go - In addition to hello, loads a logo from a gltf 2.0 file
- logo_anim.go - In addition to logo, adds some simple animation to the scene

## Troubleshooting

Check that vgelib.dll (.so) is accessible using the path.

Check that GOARCH is amd64. This is the only supported architecture.

Install the Vulkan SDK (TODO: add instructions) and see if you have any validation errors.
