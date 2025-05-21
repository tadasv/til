# FreeBSD writable live USB

FreeBSD installation images come with “Live CD” option. If you choose it, the installation image will boot in live cd mode. This is really useful if you want to explore core system from the command line or do some fixes on already installed system.

If you want to install more applications to have a better test drive you’ll be in trouble because root file system is mounted as read-only. The root file system also takes up 100% of space due to how installation image is setup leaving you no room to add new software.

There’s a very simple workaround that lets you put FreeBSD live cd into writable configuration. Boot up into usb image and enter live cd. Then run these commands:

```sh
mount -uw /
touch /firstboot
# edit (vi) /etc/fstab and change ro to rw
sysrc root_rw_mount=YES
sysrc growfs_enable=YES
```

Reboot the system when you’re done. When you boot it up next time, your root will be writable and will expand to the remaining free size of your USB stick.