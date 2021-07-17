# Xiaomi Repeater Firmware Dumper
A simple script to extract SPI flash memory contents from the Xiaomi Wi-Fi Range Extender device

An example one-liner to dump the firmware:

`go run main.go --serialname /dev/ttyUSB0 --chunksize 256 --sleep 0 --filename firmware.dump`