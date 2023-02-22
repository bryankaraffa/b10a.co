---
title: "Raspberry Pi OS Basic Setup"
date: 2023-02-20
draft: false
tags: [windows, rpi, raspberry pi, guide]
---
This is my current basic flash process for Raspberry Pi OS.

I primarely use Raspberry Pi's in a *headless* setup (no monitor or keyboard).  This process includes the basic steps of for the disk, plus a few extra steps to be able to access the machine after booting for the first time.

# 1. Flash the OS
There are plenty of existing guides for how to do this. I was using a Windows machine, so I used [Rufus](https://rufus.ie/)

You can also flash the OS using the using Raspberry Pi Imager.
Raspberry Pi Imager can install Raspberry Pi OS and other operating systems to a microSD card, ready to use with your Raspberry Pi.
You can watch the [Raspberry Pi Foundation's 45-second video](https://youtu.be/ntaXWS8Lk34) to learn how to install an operating system using Raspberry Pi Imager.

# 2. Enable SSH
Once the image is created on an SD card, the `boot` folder can be accessed on most Windows, Linux, and Mac machines by default (FAT partition). Adding certain files to this folder will activate certain setup features on the first boot of the Raspberry Pi.  For example, to **Enable SSH** server we can add:

`ssh` or `ssh.txt`

When this file is present ([RPi Documentation > Boot Folder Contents > ssh](https://www.raspberrypi.com/documentation/computers/configuration.html#ssh-or-ssh-txt)), SSH will be enabled on boot. The contents don’t matter, it can be empty. SSH is otherwise disabled by default.

Not done yet!  SSH server has been enabled, but there are no users we can use to access the machine via SSH (yet).

# 3. Create New User Account
Last step to run the RPi headless, we must create a new user account.

To do this, we can add a `userconf.txt` file to the boot folder to create a user on first boot or configure the OS with a user account using the Advanced Menu in the Raspberry Pi Imager.

This file should contain a single line of text, consisting of `username:password` – so your desired username, followed immediately by a colon, followed immediately by an encrypted representation of the password you want to use.

To generate an encrypted password, run this command on a Linux or Mac machine `openssl passwd -6`
```sh
# Example output
# Password: Super$ecret!
$ openssl passwd -6
Password: 
Verifying - Password: 
$6$dLc10VSyQ8caAzkO$nBHEJIeCbM9nOXMl840vpE8ywmRduIryQT74YTlIE6rEEmz5mqLdJhQEzB3rrbupr67xSWUBik5bneOUFWfTv0
```

The resulting `userconf.txt` looks like:
```
rpi:$6$dLc10VSyQ8caAzkO$nBHEJIeCbM9nOXMl840vpE8ywmRduIryQT74YTlIE6rEEmz5mqLdJhQEzB3rrbupr67xSWUBik5bneOUFWfTv0
```

The official docs are at [RPi Documention > Configuring a User](https://www.raspberrypi.com/documentation/computers/configuration.html#configuring-a-user)


# 4. Access RPi via SSH
```sh
# Assuming Raspberry Pi OS default hostname of `raspberrypi`
# Assuming username rpi from example above
ssh rpi@raspberrypi
rpi@raspberrypi's password:
Linux raspberrypi 5.15.61-v7+ #1579 SMP Fri Aug 26 11:10:59 BST 2022 armv7l

The programs included with the Debian GNU/Linux system are free software;
the exact distribution terms for each program are described in the
individual files in /usr/share/doc/*/copyright.

Debian GNU/Linux comes with ABSOLUTELY NO WARRANTY, to the extent
permitted by applicable law.
Last login: Tue Feb 21 05:00:12 2023 from 192.168.100.109
rpi@raspberrypi:~ $ 
```

From here, we have basic access to the RPi and can remotely update/install packages and operate the RPi.  

Depending on the use-case, we generally take further steps to [Secure the Raspberry Pi](https://www.raspberrypi.com/documentation/computers/configuration.html#securing-your-raspberry-pi)