# How to archive data to Blu-Ray M-Disc

Here's how to back-up data to BD M-Disc. This assumes we have a M-Disc
compatible rewriter.

1. Create an ISO image of your data directory that you want to burn.

    genisoimage -f -J -joliet-long -r -allow-lowercase -allow-multidot -o <image.iso> <dir>

2. Mount the image with `mount -o image.ios /mnt/cdrom` to check that all data
   and files look good. If not carefule with genisoimage flags it's possible to
   get truncated 8.3 formatted filenames.

3. Burn the image with `growisofs -speed=1 -Z /dev/sr0=image.iso`. Make sure to
   set the speed to one so that it does not mess up large BD M-Discs. Some
   people were saying that higher speed options may work but it may incorrectly
   write data to M-Disc since they like slower writes.

At this point we should have an archived copy of your data.

Burn time estimates of ~20GB iso image:

- Medium: Verbatim BD-R 25GB M-Disc
- Rewriter: LG WH16NS40 16X Super Multi M-Disc Blu-Ray BDXL DVD CD Internal
  Burner Writer Drive
- Time to burn: ~40 mins with the above speed=1 setting.