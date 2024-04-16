#!/bin/bash

# Install dependencies only if they are not installed
required_packages="build-essential libncurses-dev bison flex libssl-dev libelf-dev bc wget make gcc qemu qemu-kvm"
for package in $required_packages; do
    dpkg -s $package >/dev/null 2>&1 || {
        echo "Installing missing package: $package"
        sudo apt-get install -y $package
    }
done

# Define the working directory and the persistent download directory to avoid redownloading during testing
WORKDIR=$(pwd)/qemu_linux_boot
DOWNLOAD_DIR=$(pwd)/persistent_downloads

mkdir -p $WORKDIR
mkdir -p $DOWNLOAD_DIR
mkdir -p $WORKDIR/initramfs/{bin,sbin,etc,proc,sys,dev}

# URLs for the kernel allows for adjustment to other versions
PREFIX_NUMBER="6"
VERSION_NUMBER="6.8.6"
KERNEL_URL="https://cdn.kernel.org/pub/linux/kernel/v$PREFIX_NUMBER.x/linux-$VERSION_NUMBER.tar.xz"
KERNEL_FILE="$DOWNLOAD_DIR/kernel.tar.xz"

# Download the kernel source code if not already downloaded
[ -f $KERNEL_FILE ] || wget -O $KERNEL_FILE $KERNEL_URL

# Extract the kernel into its directory in the work directory if it hasn't been extracted
[ -d $WORKDIR/linux-$VERSION_NUMBER ] || tar -xf $KERNEL_FILE -C $WORKDIR

# Compile the kernel.
cd $WORKDIR/linux-$VERSION_NUMBER
make defconfig
make bzImage -j$(nproc)

cd $WORKDIR

# Create init program to display required message
cat > init.c <<EOF
#include <stdio.h>
#include <unistd.h>

int main() {
    printf("\\n Hello, World!\\n");
    while (1) {
        sleep(10);
    }
    return 0;
}
EOF

# Compile the init program statically and create the initramfs image
gcc -static init.c -o initramfs/init
cd initramfs
find . -print0 | cpio --null -ov --format=newc | gzip -9 > ../initramfs.img
cd ..

# Run the system in QEMU
qemu-system-x86_64 -kernel linux-$VERSION_NUMBER/arch/x86/boot/bzImage -initrd initramfs.img -append "console=ttyS0" -nographic
