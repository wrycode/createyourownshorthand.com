#include <fcntl.h>
#include <linux/input-event-codes.h>
#include <linux/input.h>
#include <linux/uinput.h>
#include <errno.h>
#include <stdio.h>
#include <unistd.h>
#include <plog/Log.h>
#include <plog/Initializers/RollingFileInitializer.h>

void emit(int fd, __u16 type, __u16 code, int val)
{
  struct input_event ie;

  ie.type = type;
  ie.code = code;
  ie.value = val;
  /* timestamp values below are ignored */
  ie.time.tv_sec = 0;
  ie.time.tv_usec = 0;

  write(fd, &ie, sizeof(ie));
}


// Helper function to press and release a key.
void press_and_release_key(int fd, __u16 code) {
   emit(fd, EV_KEY, code, 1);  // Key press
   emit(fd, EV_SYN, SYN_REPORT, 0);
   usleep(5000); // 5ms delay

   emit(fd, EV_KEY, code, 0);  // Key release
   emit(fd, EV_SYN, SYN_REPORT, 0);
   usleep(5000); // 5ms delay
}

int main()
{
  plog::init(plog::debug, "Logfile.txt"); 
  PLOGD << "main() called";
  
  struct uinput_setup usetup;

  // Open the input device
  int fd = open("/dev/uinput", O_WRONLY | O_NONBLOCK);

  // struct input_event ie;
  if(fd < 0)
    {
      perror("Failed to open /dev/uinput");
      return errno; 
    }

  /*
    * The ioctls below will enable the device that is about to be
    * created, to pass key events, in this case the space key.
    */
  ioctl(fd, UI_SET_EVBIT, EV_KEY);
  
   // Define the keys for "hello world"
   __u16 hello_world_keys[] = { KEY_H, KEY_E, KEY_L, KEY_L, KEY_O, KEY_SPACE, KEY_W, KEY_O, KEY_R, KEY_L, KEY_D };

   // Register these keys with the ioctl function
   for (size_t i = 0; i < (sizeof(hello_world_keys)/sizeof(__u16)); ++i)
   {
      ioctl(fd, UI_SET_KEYBIT, hello_world_keys[i]);
   }
  
   memset(&usetup, 0, sizeof(usetup));
   usetup.id.bustype = BUS_USB;
   usetup.id.vendor = 0x1234; /* sample vendor */
   usetup.id.product = 0x5678; /* sample product */
   strcpy(usetup.name, "Example device");

   ioctl(fd, UI_DEV_SETUP, &usetup);
   ioctl(fd, UI_DEV_CREATE);
   
   /*
    * On UI_DEV_CREATE the kernel will create the device node for this
    * device. We are inserting a pause here so that userspace has time
    * to detect, initialize the new device, and can start listening to
    * the event, otherwise it will not notice the event we are about
    * to send. This pause is only needed in our example code!
    */
   sleep(1);

   // Press and release each key
   for (size_t i = 0; i < (sizeof(hello_world_keys)/sizeof(__u16)); ++i)
   {
      press_and_release_key(fd, hello_world_keys[i]);
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
