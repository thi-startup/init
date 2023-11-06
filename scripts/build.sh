#!/bin/bash

# Define the required environment variables and their descriptions
declare -A env_vars
env_vars["drive_path"]="Drive Path"
env_vars["mount_dir"]="Mount Directory"
env_vars["init_dir"]="Init Directory"
env_vars["init_bin"]="Init Binary"
env_vars["init_config"]="Init config"

# Loop through the environment variables and check if they are set
for var_name in "${!env_vars[@]}"; do
  if [ -z "${!var_name}" ]; then
    echo "Error: Environment variable \$$var_name (${env_vars[$var_name]}) is not set."
    exit 1
  fi
done

[ -e $drive_path ] && rm -f $drive_path

fallocate -l 100M $drive_path
mkfs.ext2 $drive_path
mkdir -p $mount_dir
mount -o loop,noatime $drive_path $mount_dir
mkdir $init_dir
cp $init_bin  $init_dir
cp $init_config $init_dir
umount $mount_dir
