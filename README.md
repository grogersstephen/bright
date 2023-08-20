
### REQUIREMENTS
This program will only work on UNIX systems where the the screen brightness info can be found at '/sys/class/backlight/\<vendor\>/'
It updates the file 'brightness' to adjust the brightness

### SET UP
Install using 'go install'.

The user must be in a group which can make changes to the file at '/sys/class/backlight/\<vendor\>/brightness'
Add the user to group 'video'.
```bash
sudo usermod -aG video <user>
```
Then make sure to update the file located at:
```bash
/etc/udev/rules.d/backlight.rules
```
With the following, replace <vendor> with your vendor name, the directory found at '/sys/class/backlight/' such as 'intel_backlight':
```bash
ACTION=="add", SUBSYSTEM=="backlight", KERNEL==<vendor>, RUN+="/bin/chgrp video /sys/class/backlight/%k/brightness"
ACTION=="add", SUBSYSTEM=="backlight", KERNEL==<vendor>, RUN+="/bin/chmod g+w /sys/class/backlight/%k/brightness"
```

### USAGE
Use -t flag to set a target brightness percentage.
```bash
bright -t 50
```
Sets the brightness to 50%

-d flag will set a duration for fading. The default is 500ms.
```bash
bright -t 50 -d 3s
```

You can also use subcommands low, mid, and max.
```bash
bright low
bright mid
bright max
```
