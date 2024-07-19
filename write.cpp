// Headers for Linux input handling
#include <linux/input-event-codes.h>
#include <linux/input.h>
#include <linux/uinput.h>

// Standard C libraries for system interface
#include <fcntl.h>
#include <errno.h>
#include <stdio.h>
#include <unistd.h>

// Third-Party library headers
#include <plog/Log.h>
#include <plog/Initializers/RollingFileInitializer.h>

// C++ standard library headers
#include <locale>
#include <codecvt>
#include <string>

/* Send an event (key press/release, etc, defined in input-event-codes.h) to /dev/uinput
 * @param fd The file descriptor for the /dev/uinput device.
 * @param type The type of the event. Could be EV_KEY for key press/release events, EV_SYN to synchronize, and others
 * @param code The key code. For instance, KEY_A, KEY_ENTER, etc. 
 * @param val The value for the event. If type is EV_KEY, then 0 is for KEY_RELEASE, and 1 is for KEY_PRESS.
 */
void send_event(int fd, __u16 type, __u16 code, int val)
{
  struct input_event ie;

  ie.type = type;
  ie.code = code;
  ie.value = val;
  /* The time value isn't important when creating synthetic events.
     The kernel processes events immediately as they are generated.
  */
  ie.time.tv_sec = 0;
  ie.time.tv_usec = 0;
  write(fd, &ie, sizeof(ie));
}

// Helper function to press a key
void press_key(int fd, __u16 code) {
  send_event(fd, EV_KEY, code, 1);  // Key press
  send_event(fd, EV_SYN, SYN_REPORT, 0);
  usleep(5000); // 1ms delay
}

// Helper function to release a key
void release_key(int fd, __u16 code) {
  send_event(fd, EV_KEY, code, 0);  // Key release
  send_event(fd, EV_SYN, SYN_REPORT, 0);
  usleep(5000); // 1ms delay
}

void type_unicode(int fd, uint32_t code) {
  // Press Ctrl + Shift + U
  press_key(fd, KEY_LEFTCTRL);
  press_key(fd, KEY_LEFTSHIFT);
  press_key(fd, KEY_U);
  
  // Release keys
  release_key(fd, KEY_U);
  release_key(fd, KEY_LEFTCTRL);
  release_key(fd, KEY_LEFTSHIFT);

  // Array for mapping hexadecimal characters to keycodes
  __u16 key_map[] = {
    KEY_0, KEY_1, KEY_2, KEY_3,
    KEY_4, KEY_5, KEY_6, KEY_7,
    KEY_8, KEY_9, KEY_A, KEY_B,
    KEY_C, KEY_D, KEY_E, KEY_F
  };

  // Enter the Unicode code point
  char unicode_hex[9]; // 8 hex digits plus null terminator.
  sprintf(unicode_hex, "%08x", code);

  for (int i = 0; i < 8; i++) {
    // Get the index of the hex character to access corresponding keycode
    int index = (unicode_hex[i] <= '9') ? (unicode_hex[i] - '0') : (unicode_hex[i] - 'a' + 10);
    __u16 keycode = key_map[index];
        
    press_key(fd, keycode);
    release_key(fd, keycode);
  }

  // Press Enter to confirm
  press_key(fd, KEY_ENTER);
  release_key(fd, KEY_ENTER);
}

/* Some background reading:
   UTF-8 explained simply: https://www.tsmean.com/articles/encoding/unicode-and-utf-8-tutorial-for-dummies/
   
   TODO: better document this function
 */
std::vector<uint32_t> convertToUnicode(std::u8string const& str)
{
    std::vector<uint32_t> unicode_values;
    char32_t ch = 0;
    int bytes_in_ch = 0;

    for (unsigned char c : str)
    {
        if (c <= 0x7F) // 1-byte/7-bit ascii
            ch = c, bytes_in_ch = 1;
        else if ((c & 0xE0) == 0xC0) // 2-byte/11-bit in 5+6 format
            ch = c & 0x1F, bytes_in_ch = 2;
        else if ((c & 0xF0) == 0xE0) // 3-byte/16-bit in 4+6+6 format
            ch = c & 0x0F, bytes_in_ch = 3;
        else if ((c & 0xF8) == 0xF0) // 4-byte/21-bit in 3+6+6+6 format
            ch = c & 0x07, bytes_in_ch = 4;
        else // continuation bytes
            ch = (ch << 6) | (c & 0x3F);
        if (--bytes_in_ch <= 0)
            unicode_values.push_back(ch), bytes_in_ch = 0;
    }
    return unicode_values;
}

int main()
{

  // TODO, set up error logging to logfile+stderr so we don't use perror
  plog::init(plog::debug, "Logfile.txt"); 
  PLOGD << "main() called";

  // request a virtual device from the uinput kernel subsystem; see https://kernel.org/doc/html/v4.12/input/uinput.html
  int fd = open("/dev/uinput", O_WRONLY | O_NONBLOCK);
  if(fd < 0)
    {
      PLOGE << "Failed to open /dev/uinput";      
      perror("Failed to open /dev/uinput");
      return errno; 
    }

  // enable key press/release events
  ioctl(fd, UI_SET_EVBIT, EV_KEY);

  // enable other needed keys
  __u16 unicode_keys[] = { KEY_LEFTCTRL, KEY_LEFTSHIFT, KEY_U, KEY_ENTER, KEY_A, KEY_B, KEY_C, KEY_D, KEY_E, KEY_F,
                           KEY_0, KEY_1, KEY_2, KEY_3, KEY_4, KEY_5, KEY_6, KEY_7, KEY_8, KEY_9 };
  for (size_t i = 0; i < (sizeof(unicode_keys)/sizeof(__u16)); ++i)
    {
      ioctl(fd, UI_SET_KEYBIT, unicode_keys[i]);
    }

  // TODO move all this virtual keyboard device setup and creation code into its own function
  struct uinput_setup usetup;

  memset(&usetup, 0, sizeof(usetup));
  usetup.id.bustype = BUS_USB;
  usetup.id.vendor = 0x1234; /* sample vendor */
  usetup.id.product = 0x5678; /* sample product */
  strcpy(usetup.name, "Example device");

  ioctl(fd, UI_DEV_SETUP, &usetup);
  ioctl(fd, UI_DEV_CREATE);

  PLOGI << "Successfully initialized virtual keyboard";
  /*
   * On UI_DEV_CREATE the kernel will create the device node for this
   * device. We are inserting a pause here so that userspace has time
   * to detect, initialize the new device, and can start listening to
   * the event, otherwise it will not notice the event we are about
   * to send. This pause is only needed in our example code!
   */
  sleep(1);

  // The correct way to type special characters manually is
  // Ctrl+Shift+u (then release) followed by the hexadecimal Unicode
  // code point (e.g. 2 + 0 + b + 9) + Enter or Space
  std::u8string str = u8"Russian: Привет, мир! Это тестовый текст.(Hello, world! This is test text.)\nArabic: مرحبا بك في العالم! هذا نص تجريبي. (Hello, world! This is test text.)";
  std::vector<uint32_t> code_points = convertToUnicode(str);
  
  for (size_t i = 0; i < code_points.size(); ++i)
    {
      type_unicode(fd, code_points[i]);
    }

  /*
   * Give userspace some time to read the events before we destroy the
   * device with UI_DEV_DESTROY.
   */
  sleep(1);

  ioctl(fd, UI_DEV_DESTROY);

  close(fd);

  return 0;
}

// std::string
