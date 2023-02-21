---
title: "Using APRSDroid with Baofeng UV-5R"
date: 2023-01-14
draft: true
---

I have recently been interested in getting into amateur radio.  There are a lot of aspects to amateur radio, but one thing that peaked by interest was APRS (Automatic Packet Reporting System).  

> APRS is a digital communications protocol that allows for the transmission of data over radio frequencies.  APRS is used for a variety of things, including weather reporting, traffic reporting, and tracking of vehicles and people.

In the US, if you tune to `144.390MHz` you can hear what APRS traffic sounds like.  The brief, ~1second transmissions you should hear periodically are APRS "packets", which can be converted from the analog sound signal to the digital data.

## Getting started

I wanted to get started transmitting location via APRS, but I didn't want to spend a lot of money on equipment.  I found an app called [APRSDroid](https://github.com/ge0rg/aprsdroid) which allows you to use an Android phone to read/send APRS packets in conjuction with a radio.  The app costs a few dollars from the Google Play store, but it is also possible to [download the APK](https://aprsdroid.org/download/) for manual installation which is free.

Here's what I used:
 - A Radio Receiver {{< amzn-affiliate text="Baofeng UV-5RA" link="https://amzn.to/3W9CrVu" >}}
 - An Audio Interface Cable {{< amzn-affiliate text="BTECH APRS-K1 Cable" link="https://amzn.to/3GIFH4u" >}}
 - An Android device that has a headphone jack (Moto G6)

## Setting up the radio

### APRSDroid
First, install the APRSDroid on your phone.  You can download the APK from the [APRSDroid website](https://aprsdroid.org/download/).  I installed the app on my Moto G6, but it should work on any Android device that has a headphone jack.

After install, open the app and go to the "Settings" tab.  You will need to configure the app to use "AFSK".

### Audio Interface Cable
Next, connect the audio interface cable to the radio.  When I plugged the cable into the headphone jack on the Phone, it detected a "mic and headset" and the icon appeared, which is what we want.  The cable is basically connecting:
  - Radio's audio output (radio speaker) **->** Phone's audio input [phone mic]
  - Radio's audio input (radio mic) **<-** Phone's audio output [phone headset]