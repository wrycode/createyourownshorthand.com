git clone --recurse-submodules --shallow-submodules https://github.com/wrycode/vtt.git

sudo pacman -Sy abseil-cpp grpc curl nlohmann-json, probably a few more I forgot

AUR packages: google-crc32c

git submodule update --init --recursive

cd extern/google-cloud-cpp
cmake -S . -B cmake-out \
-DBUILD_TESTING=OFF \
-DGOOGLE_CLOUD_CPP_ENABLE_EXAMPLES=OFF \
-DGOOGLE_CLOUD_CPP_ENABLE=speech \
-DCMAKE_INSTALL_PREFIX=../

cmake --build cmake-out -- -j "$(nproc)"
cmake --build cmake-out --target install

export GOOGLE_APPLICATION_CREDENTIALS=~/.config/gcloud/application_default_credentials.json




sudo vim /etc/udev/rules.d/uinput.rules:
KERNEL=="uinput", GROUP="vtt", MODE="0660"
sudo udevadm control --reload-rules && sudo udevadm trigger

vtt $ sudo groupadd vtt
vtt $ sudo usermod -a -G vtt wrycode

sudo    setfacl -m g::rw /dev/uinput

Or just give up:
sudo chmod u+w /dev/uinput 
