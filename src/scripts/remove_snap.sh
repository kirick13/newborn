#!/bin/sh

while [ "$(snap list | wc -l)" -gt 1 ]; do
    removed_any=0
    for snap in $(snap list | awk 'NR>1 {print $1}'); do
        if snap remove "$snap"; then
            removed_any=1
        fi
    done
    [ "$removed_any" -eq 1 ] || (echo 'Failed to remove snaps' && exit 1)
done
