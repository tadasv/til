# Virtual monitors with xrandr

xrandr offers powerful screen management capabilities. You can slice single monitor into several virtual screens. Below script is used
to split single 4k screen into 3:

```
+----+----+
| ~1 |    |
+----+ ~3 |
| ~2 |    |
+----+----+
```

Save this to virtmon.sh and use it to toggle between virtual screen and original 4k.

```sh
#!/bin/sh

if [ -z "$(xrandr --listactivemonitors | grep 'DisplayPort-2~1')" ]; then
	xrandr --setmonitor DisplayPort-2~1 1920/470x1080/264+0+0 DisplayPort-2
	xrandr --setmonitor DisplayPort-2~2 1920/470x1080/264+0+1080 none
	xrandr --setmonitor DisplayPort-2~3 1920/470x2160/264+1920+0 none
	xrandr --fb 3840x2160
else
	xrandr --delmonitor DisplayPort-2~1
	xrandr --delmonitor DisplayPort-2~2
	xrandr --delmonitor DisplayPort-2~3
fi
```
