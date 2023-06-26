#!/bin/bash
BUILDDIR=/mnt/arm-build
# Prepare the attached disk, this should probably be automated to get the attached
# device, then mount it based on an env var for the mnt location
DEVICE=$(readlink -f /dev/disk/by-id/google-persistent-disk-1) 
parted $DEVICE mklabel gpt 
parted $DEVICE mkpart primary ext4 0% 100% 
PART=${DEVICE}1 
sleep 5
mkfs.ext4 $PART 
mkdir $BUILDDIR 
mount $PART $BUILDDIR

# We should try to get username from gcloud or the gcp api based on the ssh keys
chown -R sheran:sheran $BUILDDIR
echo "$PART $BUILDDIR ext4 defaults 0 0" | tee -a /etc/fstab

apt update
apt install -y git python3 build-essential crossbuild-essential-arm64 screen jq binutils-aarch64-linux-gnu

# Install tailscale
curl -fsSL https://tailscale.com/install.sh | sh
tailscale up --authkey={{.TsKey}}