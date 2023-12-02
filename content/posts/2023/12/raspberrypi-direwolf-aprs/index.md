---
title: "Using Raspberry Pi, RTL-SDR, and Direwolf for cheap APRS receiver"
date: 2023-12-01
tags: [amateur radio, aprs, open source, baofeng, raspberry pi, rtl-sdr, direwolf]
---

## Install RTL-SDR and Direwolf
```sh
sudo apt-get update -y && sudo apt-get upgrade -y

sudo apt-get install -y rtl-sdr direwolf

sudo reboot now
```

## Test RTL-SDR
```sh
rtl_test
```
{{< figure src="Screenshot 2023-02-26 204405.png" alt="rtl_test output" >}}

## Test Receiving APRS with Direwolf
```sh
rtl_fm -f 144.39M - | direwolf -r 24000 -D 1 -
```

{{< figure src="Screenshot 2023-02-26 205519.png" alt="rtl_fm output piped to direwolf" >}}

## Starting Direwolf on boot
We need to create a systemd service configuration so the services start on boot.  We should also redirect the stdout/stderr outputs to syslog for debugging in the future.

`/usr/local/bin/direwolf-aprs.sh`:
```sh
#!/bin/bash

/usr/bin/rtl_fm -f 144.39M - | /usr/bin/direwolf -r 24000 -D 1 -
```

Make `direwolf-aprs.sh` executable
```sh
sudo chmod +x /usr/local/bin/direwolf-aprs.sh
```

Allow direwolf user to use rtl-sdr USB device
```sh
sudo wget -O /etc/udev/rules.d/rtl-sdr.rules "https://raw.githubusercontent.com/osmocom/rtl-sdr/master/rtl-sdr.rules"

sudo systemctl restart udev.service

sudo usermod -a -G plugdev direwolf
```

`/lib/systemd/system/direwolf.service`:
```toml
[Unit]
Description=DireWolf is a software "soundcard" modem/TNC and APRS decoder
Documentation=man:direwolf

[Service]
User=direwolf
SupplementaryGroups=dialout
ExecStart=/usr/local/bin/direwolf-aprs.sh
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=direwolf

[Install]
WantedBy=multi-user.target
```

Reload systemctl:
```sh
sudo systemctl daemon-reload
```

Enable the direwolf service:
```sh
sudo systemctl enable direwolf
```

Start the direwolf service:
```sh
sudo systemctl start direwolf
```

Validate direwolf service is running:
```sh
sudo systemctl status direwolf
```

*For additional reading, a more in-depth guide for this process is the [Raspberry Pi SDR IGate by WB20SZ](https://github.com/wb2osz/direwolf/blob/master/doc/Raspberry-Pi-SDR-IGate.pdf)*

## Optional: Adjust audio levels
To be able to decode packets received by your station efficiently, you will need to adjust the volume of the incoming audio so it's around 50%. This is done by adjusting the volume of the Speaker in `alsamixer`. I typically set it to 50 on Raspberry Pi's default device.

Check audio levels are good using direwolf logs
```sh
sudo tail -f /var/log/syslog | grep direwolf

# Example output:
# Digipeater WIDE2 (probably K6TZ-9) audio level = 47(12/17)   [NONE]   ||||||___`
```
**audio level = 47** means 47% (which is good)

Adjust the audio levels with `alsamixer`
```sh
alsamixer
```
`Esc` to Exit.

Save alsa settings
```sh
sudo alsactl store
```

## Optional: Enable logrotate
Because we are saving the output from direwolf to syslog [typically `/var/log/syslog`], this can lead to the root volume disk becoming full.  We can enable `logrotate` to help mitigate and prevent any issues related to disk usage from the common log files.

```sh
sudo systemctl enable logrotate
sudo systemctl start logrotate
```

## Connect to direwolf with APRS Client (using AGW protocol)
I use [APRSIS32](http://aprsisce.wikidot.com/) on Windows, and [YAAC](http://www.ka2ddo.org/ka2ddo/YAAC.html) which works on everything that has Java instdalled -- b[ut any client supporting AGW network Interface (TCP/IP) protocol can talk to direwolf](https://github.com/wb2osz/direwolf#dire-wolf-includes).


Here's an example of APRSIS32 connected to direwolf AGW endpoint listening on port `8080` {{< figure src="Screenshot 2023-02-28 205942.png" alt="APRSIS32 on Windows connected to direwolf on Raspberry Pi" >}}
