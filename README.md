# Build instructions  * 
1. git clone --recurse-submodules --shallow-submodules https://github.com/wrycode/vtt.git

2. Install development dependencies: CMake, GCC, [just](https://github.com/casey/just), probably a few more I've forgotten

3. Install dependencies for the Cloud Speech-to-Text API C++ Client Library: https://github.com/googleapis/google-cloud-cpp/blob/main/doc/packaging.md#required-libraries

For instance, for me, I ran: `sudo pacman -Sy cmake abseil-cpp grpc curl nlohmann-json just`, plus installed  [google-crc32c](https://aur.archlinux.org/packages/google-crc32c) from the AUR

4. Commands to build and run different components are in flux and documented in [.justfile](.justfile)


# Enabling virtual keyboard

Keys are typed using the [uinput](https://kernel.org/doc/html/v4.12/input/uinput.html) kernel interface, exposed in /dev/uinput. The program needs to have permissions to write to /dev/uinput. The easiest way is to chmod /dev/uinput, but there is probably a smarter way with udev.


# Unicode characters

For languages other than English, you must have [IBus](https://wiki.archlinux.org/title/IBus). It may be enabled/installed already. You can test if it is installed by typing Ctrl+Shift+U (then release), then 2+7+0+5. You should see: ✅
