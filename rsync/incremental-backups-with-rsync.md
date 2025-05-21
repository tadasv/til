# Incremental backups with rsync

Inspired by [https://linuxconfig.org/how-to-create-incremental-backups-using-rsync-on-linux](https://linuxconfig.org/how-to-create-incremental-backups-using-rsync-on-linux)

This script will create incremental backups using rsync. This is achieved by
using hard links support of the filesystem.

```sh
#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

readonly SOURCE_DIR="/home/vptr"
readonly BACKUP_DIR="/mnt/backup/vptr"
readonly DATETIME="$(date '+%Y-%m-%d_%H:%M:%S')"
readonly BACKUP_PATH="${BACKUP_DIR}/${DATETIME}"
readonly LATEST_LINK="${BACKUP_DIR}/latest"

umount /mnt/backup || true
mount /mnt/backup

mkdir -p "${BACKUP_DIR}"

rsync -av --delete \
    "${SOURCE_DIR}/" \
    --link-dest "${LATEST_LINK}" \
    --exclude=".cache" \
    --exclude="*/node_modules/*" \
    --exclude="Downloads*" \
    "${BACKUP_PATH}"

rm -rf "${LATEST_LINK}"
ln -s "${BACKUP_PATH}" "${LATEST_LINK}"

umount /mnt/backup
```