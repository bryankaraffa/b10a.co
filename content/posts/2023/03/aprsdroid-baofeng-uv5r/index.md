---
title: "Using APRSDroid with Baofeng UV-5R"
date: 2023-03-01
draft: false
tags: [amateur radio, aprs, open source, baofeng, cheap, guide]
---

I have recently been interested in getting into amateur radio.  There are a lot of aspects to amateur radio, but one thing that peaked by interest was APRS (Automatic Packet Reporting System).

> **APRS** stands for Automatic Packet Reporting System, a digital communication system used by radio amateurs. It can be used to send your position, share telemetry (a weather station for example), and even send text messages to other hams.

In the US, if you tune to `144.390MHz` you can hear what APRS traffic sounds like.  The brief, ~1second transmissions you should hear periodically are APRS "packets", which can be converted from the analog sound signal to the digital data.

> **Baofeng UV-5R** is an inexpensive Chinese dual-band handheld radio that can be used for APRS with some additional hardware and software

> **APRS Droid** is an Android application that can be used to encode and decode APRS packets using your smartphone’s GPS and audio jack.

To set up Baofeng UV-5R and APRS Droid for APRS communication, you will need:

 - A Baofeng UV-5R radio {{< amzn-affiliate text="Baofeng UV-5RA" link="https://amzn.to/3W9CrVu" >}}
 - A cable to connect the radio to your smartphone’s audio jack (such as Btech APRS cable) {{< amzn-affiliate text="BTECH APRS-K1 Cable" link="https://amzn.to/3GIFH4u" >}}
 - An Android smartphone with APRS Droid installed. The app costs a few dollars from the Google Play store, but it is also possible to [download the APK](https://aprsdroid.org/download/) for manual installation which is free.
 - A valid amateur radio license


## Setting up the radio
You will need to program your radio to use the appropriate frequency and settings for APRS. The most common frequency for APRS in North America is 144.390 MHz . You can use a software such as CHIRP to program your radio using a computer and a programming cable, or you can do it manually using the keypad on the radio.

To program your radio manually, follow these steps:

Turn on your radio and press VFO/MR to enter frequency mode
 - Enter 144.390 using the keypad
 - Press A/B to select channel A
 - Press MENU and use the arrow keys to select VOX
 - Press MENU again and enter 10 using the keypad (this sets the VOX level to maximum)
 - Press MENU again to confirm and exit

You have now programmed your radio for APRS. You can verify that it is working by listening for APRS packets on `144.390 MHz`. You should hear short bursts of data every few seconds.

## Setting up APRSDroid
These steps install the APRSDroid on your phone.  You can download the APK from the [APRSDroid website](https://aprsdroid.org/download/) which is free, or install from the Google Play store which is a small cost, but helps support the developers.  I installed the app via APK on my Moto G6, and it should work on any Android device that has a headphone jack.

After installing the app, you can set up your APRS Droid app, follow these steps:

Open the app and tap on Preferences
 - Tap on Connection Preferences and select AFSK (via mic/speaker)
 - Tap on Callsign and enter your callsign and passcode
 - Tap on APRS Symbol and choose a symbol that represents you (such as a car or a person)
 - Tap on SmartBeaconing and enable it (this will adjust your beacon rate based on your speed and direction)
 - Tap on Back to return to the main screen

You have now configured your APRS Droid app. You can verify that it is working by tapping on Start Tracking or Send Location and you should hear a short burst and see green line appear confirming "ASFK OK"

## Connecting Audio Interface Cable
You will need a cable that connects the speaker and microphone jacks of your radio to the audio jack of your smartphone. These are readily available for purchase online, such as Btech APRS cable ({{< amzn-affiliate text="BTECH APRS-K1 Cable" link="https://amzn.to/3GIFH4u" >}}).

When I plugged the cable into the headphone jack on the Phone, it detected a "mic and headset" and the icon appeared, which is what we want.  The cable is basically connecting:
  - Radio's audio output (radio speaker) **->** Phone's audio input [phone mic]
  - Radio's audio input (radio mic) **<-** Phone's audio output [phone headset]

To connect your radio and your smartphone, follow these steps:
 - Plug one end of the cable into the speaker and microphone jacks of your radio
 - Plug the other end of the cable into the audio jack of your smartphone
 - Turn on your radio and set it to channel A (144.390 MHz)
 - Turn on your smartphone and open APRS Droid app
 - Tap on Start Tracking to begin sending and receiving APRS packets

## Watching it work
If everything is setup correctly, and you are within range of APRS receivers (iGates or Digipeaters).

APRS was originally designed to be a local, tactical, real-time two-way communications system -- today with the help from APRS Internet Gateways (iGates) -- all local information is injected into the APRS Internet Service: https://aprs.fi/

