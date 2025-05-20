# Disable wifi power management

Power management might interfere with wifi speeds

```sh
# install tools
$ apt install wireless-tools
# turn off power management
$ iwconfig wlan0 power off
```

After module is installed/updated, you can check whether power saving settings
are on or not

```sh
$ cat /sys/module/iwlwifi/parameters/power_save
```

The power still can be limited by iwcl or other devices since after reboot
iwconfig shows power management on, but parameters file above shows it off.

Debian wiki has a page on Intel wifi chipsets https://wiki.debian.org/iwlwifi

To improve performance of AX210

```
#/etc/modprobe.d/iwlwifi.conf
options iwlwifi bt_coex_active=0 swcrypto=1 11n_disable=8
```

```
#/etc/modprobe.d/iwlmvm.conf
options iwlmvm power_scheme=1
```

These settings should fix driver crashes.