# X11 clipboard synchronization

Clipboard in Linux was always a PITA as far as I can remember. We have three clipboards to work with in X11: primary, secondary and clipboard. Primary buffer is typically used in terminal apps (e.g. when selecting text with mouse), secondary not sure, and the “clipboard” is used by GUI applications such as you web browser and such.

These clipboards are not automatically synced by default. For example, if you select console text and try to paste it in your browser it won’t work out of the box. You can transfer primary buffer context to clipboard with xclip like so:

```sh
$ xclip -o | xclip -selection c
```

Then it will work. But who wants to do this all the time.

There’s a program called `autocutsel` that can help you with clipboard sync. You just need to install it and add the following to `~/.xinitrc`:

```sh
if [ -x /usr/bin/autocutsel ]; then
    # keep clipboard in sync with primary buffer
    autocutsel -selection CLIPBOARD -fork
    # keep primary buffer in sync with clipboard
    autocutsel -selection PRIMARY -fork
fi
```

It will start autocutsel in the background when X starts and keep clipboards in sync.