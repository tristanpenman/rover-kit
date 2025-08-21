# ROS Notes

Source installation on Raspbian
  - instructions at http://wiki.ros.org/kinetic/Installation/Source

Pre-built armhf packages not yet available for Debian Jessie
  - see http://answers.ros.org/question/239466/incomplete-packages-for-kinetic-armhf-jessie/

Alternative is to use Ubuntu Xenial base (e.g. Ubuntu Mate)
  - instructions at http://wiki.ros.org/indigo/Installation/UbuntuARM

## Qemu

Download the kernel from https://github.com/dhruvvyas90/qemu-rpi-kernel

Expand the Raspbian image:

    qemu-img resize -f raw 2017-04-10-raspbian-jessie-lite.img +5G

Before starting the virtual machine, change the interrupt sequence from CTRL-C (`^C`) to CTRL-[ (`^[`):

    stty intr ^]

Start the virtual machine:

    qemu-system-arm \
      -cpu arm1176 \
      -m 256 \
      -M versatilepb \
      -no-reboot \
      -serial stdio \
      -append "root=/dev/sda2 panic=1 rootfstype=ext4 rw systemd.log_target=null earlyprintk=serial loglevel=8 console=ttyAMA0,115200" \
      -kernel qemu-rpi-kernel-master/kernel-qemu-4.4.34-jessie \
      -drive "file=2017-04-10-raspbian-jessie-lite.img,index=0,media=disk,format=raw"

Experimental pi 2 support:

    qemu-system-arm \
      -M raspi2 \
      -serial stdio \
      -dtb bcm2709-rpi-2-b.dtb \
      -kernel kernel7.img \
      -append "dwc_otg.fiq_fix_enable=0 root=/dev/mmcblk0p2 panic=1 rootfstype=ext4 rw earlyprintk loglevel=8 console=ttyAMA0,115200" \
      -drive "file=2017-04-10-raspbian-jessie-lite-frsh.img,format=raw,if=sd"

Should be possible to deploy this image directly to the Raspberry Pi.

## WPA supplicant

Example /etc/wpa_supplicant/wpa_supplicant.conf:

    ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
    update_config=1

    network{
       ssid="Moonlight"
       psk="<My password>"
    }

Then:

    sudo ifdown wlan0
    sudo ifup wlan0

## Setup ROS Repositories

    sudo sh -c 'echo "deb http://packages.ros.org/ros/ubuntu $(lsb_release -sc) main" > /etc/apt/sources.list.d/ros-latest.list'
    sudo apt-key adv --keyserver hkp://ha.pool.sks-keyservers.net:80 --recv-key 421C365BD9FF1F717815A3895523BAEEB01FA116
    sudo apt-get update
    sudo apt-get upgrade

## Install boostrap dependencies

    sudo apt-get install -y python-rosdep python-rosinstall-generator python-wstool python-rosinstall build-essential cmake

## Initialize rosdep

    sudo rosdep init
    rosdep update

## Create a ros_catkin Workspace

First, we need to build console-bridge using Boost 1.55 to avoid warnings later:

    sudo sh -c 'echo "deb-src http://mirrordirector.raspbian.org/raspbian/ testing main contrib non-free rpi" >> /etc/apt/sources.list'
    sudo apt-get update
    sudo apt-get build-dep console-bridge
    mkdir -p ~/ros_catkin_ws/external_src
    cd ~/ros_catkin_ws/external_src
    apt-get source -b console-bridge
    sudo dpkg -i libconsole-bridge0.2*.deb libconsole-bridge-dev_*.deb

For reference, this is the warning that we're avoiding (noticed while building ros_tutorials package):

    "/usr/bin/ld: warning: libboost_system.so.1.54.0, needed by /usr/lib/gcc/arm-linux-gnueabihf/4.9/../../../arm-linux-gnueabihf/libconsole_bridge.so, may conflict with libboost_system.so.1.55.0"

Now we can generate the ROS installation config:

    cd ~/ros_catkin_ws
    rosinstall_generator ros_comm --rosdistro kinetic --deps --wet-only --tar > kinetic-ros_comm-wet.rosinstall
    wstool init src kinetic-ros_comm-wet.rosinstall

    cd ~/ros_catkin_ws
    rosdep install -y --from-paths src --ignore-src --rosdistro kinetic -r --os=debian:jessie
    sudo ./src/catkin/bin/catkin_make_isolated --install -DCMAKE_BUILD_TYPE=Release --install-space /opt/ros/kinetic -j2

## Adding packages to workspace

Example is `sensor_msgs` package, required by Matt's ROS node wrapper for hc-sr04 sensors:

    cd ~/ros_catkin_ws
    rosinstall_generator ros_comm sensor_msgs --rosdistro kinetic --deps --wet-only --tar > kinetic-custom_ros.rosinstall
    wstool merge -t src kinetic-custom_ros.rosinstall
    wstool update -t src

As above:

    rosdep install --from-paths src --ignore-src --rosdistro kinetic -y -r --os=debian:jessie
    sudo ./src/catkin/bin/catkin_make_isolated --install -DCMAKE_BUILD_TYPE=Release --install-space /opt/ros/kinetic -j2

## After logins

    source /opt/ros/kinetic/setup.bash

## Catkin Tutorial Workspace

    mkdir -p ~/catkin_ws/src
    catkin_init_workspace ~/catkin_ws/src
    cd ~/catkin_ws
    catkin_make
    cd ~/catkin_ws/src
    catkin_create_pkg beginner_tutorials std_msgs rospy roscpp

## Matt's ROS node wrapper for hc-sr04

Dependencies (other than `sensor_msgs` package):

    sudo apt-get install wiringpi

    mkdir -p ~/rover_catkin_ws/src
    cd ~/rover_catkin_ws/src
    git clone https://github.com/matpalm/ros-hc-sr04-node.git
    cd ..
    catkin_make

## Source installation on Ubuntu Desktop

    sudo sh -c 'echo "deb http://packages.ros.org/ros/ubuntu $(lsb_release -sc) main" > /etc/apt/sources.list.d/ros-latest.list'
    sudo apt-key adv --keyserver hkp://pgp.mit.edu:80 --recv-key 421C365BD9FF1F717815A3895523BAEEB01FA116
    sudo apt-get update
    sudo apt-get install python-rosdep python-rosinstall-generator python-wstool python-rosinstall build-essential
    sudo rosdep init
    mkdir -p ~/ros_catkin_ws/src
    cd ~/ros_catkin_ws
    rosinstall_generator desktop --rosdistro kinetic --deps --tar > kinetic-desktop.rosinstall
    wstool init -j8 src kinetic-desktop.rosinstall
    rosdep install --from-paths src --ignore-src --rosdistro kinetic -y
    ./src/catkin/bin/catkin_make_isolated --install -DCMAKE_BUILD_TYPE=Release

    source ~/ros_catkin_ws/install_isolated/setup.bash
