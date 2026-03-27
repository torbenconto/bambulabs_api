# Target Printer Behavior
This document outlines the behavior expected of bambulabs 3d printers by this library.

## Purpose
The purpose of this document is to describe the current capabilities of the library and the target behavior it expects. This applies to both the library and the emulator package. If at any point the behavior described in this file does not line up with real printer behavior, the library and emulator are very likely out of date.

## General
From now on, we will assume that the printer has a running MQTT and FTP service on the normal ports with normal credentials as outlined [here](https://github.com/Doridian/OpenBambuAPI/blob/main/mqtt.md).

### X1 series printers
These are the printers that I have the most readily avalible access to so this will be the most accurate.

I have observed MQTT pushall updates to be pushed from the printer unsolicited every 5-10 seconds. This is the only behavior really specific to these higher end printers, everything else like the camere being RTSPS only is outlined in openbambuapi.
