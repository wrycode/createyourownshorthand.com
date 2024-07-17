git clone --recurse-submodules <project-url>
git submodule update --init --recursive





sudo vim /etc/udev/rules.d/uinput.rules:
KERNEL=="uinput", GROUP="vtt", MODE="0660"
sudo udevadm control --reload-rules && sudo udevadm trigger

vtt $ sudo groupadd vtt
vtt $ sudo usermod -a -G vtt wrycode

sudo    setfacl -m g::rw /dev/uinput

Or just give up:
sudo chmod u+w /dev/uinput 
