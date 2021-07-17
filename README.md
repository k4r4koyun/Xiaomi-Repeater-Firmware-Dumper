# Xiaomi Repeater Firmware Dumper
A simple script to extract SPI flash memory contents from the Xiaomi Wi-Fi Range Extender device

![FirmwareToolOutput](https://user-images.githubusercontent.com/9578833/126046423-a119f8a3-f7d2-45bf-98af-a085ba8f45e6.png)

An example one-liner to dump the firmware:

`go run main.go --serialname /dev/ttyUSB0 --chunksize 256 --sleep 0 --filename firmware.dump`