# Allow 3rd party binaries on Mac OS

When running 3rd party unsigned binaries for Mac, you will be welcomed with a dialog
letting you know that you cannot run this app because it's not verified. This requires the user
going to the Privacy & Settings and manually allowing the app.

It's ok for a single app, but problematic for cases where you have lots of binaries/libraries. It seems that
any newly downloaded binaries are added to quarantine by Mac OS, which triggers Gatekeeper. A workaround is
to remove quarantine attribute from the files you downloaded:

```sh
$ xattr -d com.apple.quarantine *
```