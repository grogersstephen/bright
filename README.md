
### REQUIREMENTS
This program will only work on UNIX systems where the the screen brightness info can be found at '/sys/class/backlight/\<vendor\>/'
It updates the file 'brightness' to adjust the brightness

### SET UP
Install using 'go install'.

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
