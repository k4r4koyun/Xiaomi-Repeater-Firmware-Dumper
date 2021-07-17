package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sbwhitecap/tqdm"
	"github.com/tarm/serial"
)

func travelToSPIMenu(serialPort *serial.Port) {
	// Travel to the top level menu, if not already there
	_, err := serialPort.Write([]byte(";up;up;os;"))
	if err != nil {
		log.Fatalln("Could not write to serial port:", err)
	}

	buf := make([]byte, 2048)

	// Clean the UART buffer
	_, err = serialPort.Read(buf)
	if err != nil {
		log.Fatalln("Could not read serial port:", err)
	}
}

func readNBytes(serialPort *serial.Port, byteCount int, offset int) []byte {
	var err error

	// Format the SPI memory read command, EX: "spi rd 0x0 64"
	writeString := "spi rd " + fmt.Sprintf("0x%08X", offset) + " " + strconv.Itoa(byteCount)

	_, err = serialPort.Write([]byte(writeString + ";"))
	if err != nil {
		log.Fatalln("Could not write to serial port:", err)
	}

	currentChunk := ""

	// Run read cycles until fragment dump is complete
	tempBuf := make([]byte, byteCount*2+63)
	for !strings.Contains(string(tempBuf), "\r\n\r\n\r\n") {
		// tempBuf = tempBuf[:0]

		n, err := serialPort.Read(tempBuf)
		if err != nil {
			log.Fatalln("Could not read serial port:", err)
		}

		currentChunk += string(tempBuf[:n])
	}

	// Find the index of the read command, as it is echoed back
	cmdIndex := strings.Index(currentChunk, writeString)
	if cmdIndex == -1 {
		log.Println(currentChunk)
		log.Fatalln("Could not find delimiter, please plug the device again to manually clear UART buffer and wait for the device to complete boot process")
	}
	cmdIndex += len(writeString) + 2

	// Extract the hex representation from response
	chunkLength := (byteCount * 2) + (byteCount / 4) + (byteCount / 8)
	finalChunk := currentChunk[cmdIndex : cmdIndex+chunkLength]

	// Tidy up the string representation of hex chunk
	finalChunk = strings.ReplaceAll(finalChunk, "\r", "")
	finalChunk = strings.ReplaceAll(finalChunk, "\n", "")
	finalChunk = strings.TrimSpace(finalChunk)

	// Reverse the byte order, 4 bytes at a time
	encodedString := ""
	fourByteSlice := strings.Fields(finalChunk)
	for _, v := range fourByteSlice {
		encodedString += v[6:8] + v[4:6] + v[2:4] + v[0:2]
	}

	// Turn string representation into real hex values
	decodedBytes, err := hex.DecodeString(encodedString)
	if err != nil {
		log.Fatalln("Could not decode the hex string:", err)
	}

	return decodedBytes
}

func main() {
	var chunkSize, cycleSleep int
	var fileName, serialName string

	// Prepare command line flags and parse them
	flag.IntVar(&chunkSize, "chunksize", 512, "Chunk size for firmware fragments [64-128-256-512-1024]")
	flag.IntVar(&cycleSleep, "sleep", 0, "Amount to sleep between fragmented reads [Max 100 ms]")
	flag.StringVar(&fileName, "filename", "firmware.dump", "File name for the dumped firmware")
	flag.StringVar(&serialName, "serialname", "", "Serial port name for the connected device") // Required
	flag.Parse()

	// Check chunk size parameter boundaries
	switch chunkSize {
	case 64:
		break
	case 128:
		break
	case 256:
		break
	case 512:
		break
	case 1024:
		break
	default:
		log.Fatalln("Wrong chunk size, must be one of [64-128-256-512-1024]")
	}

	// Check if sleep value is provided correctly
	if cycleSleep < 0 || cycleSleep > 100 {
		log.Fatalln("That much sleep is not good for the body")
	}

	// Check for serial name, could be something like "/dev/ttyUSB0" or "COM1"
	if serialName == "" {
		log.Fatalln("Serial port name should be specified!")
	}

	fmt.Println()
	fmt.Println("======= XIAOMI WI-FI REPEATER - FIRMWARE DUMP TOOL =======")
	fmt.Println()
	fmt.Println("= Processor:", "MediaTek MT7628KN")
	fmt.Println("= Memory:", "Macronix MX25L1606E")
	fmt.Println("= Chunk Size:", chunkSize, "Bytes")
	fmt.Println("= Cycle Sleep:", cycleSleep, "Milliseconds")
	fmt.Println("= Filename:", fileName)
	fmt.Println("= Serial Port Name:", serialName)
	fmt.Println("= Memory Size:", 2097152, "Bytes")
	fmt.Println("= Total Read Cycles:", 2097152/chunkSize)
	fmt.Println()

	// Initialize serial port connection
	serialConfig := &serial.Config{Name: serialName, Baud: 115200}
	serialPort, err := serial.OpenPort(serialConfig)
	if err != nil {
		log.Fatal("Could not create serial port connection: ", err)
	}

	travelToSPIMenu(serialPort)

	// Open firmware dump file
	dumpFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalln("File open error: ", err)
	}
	defer dumpFile.Close()

	index := 0
	iterSize := 2097152 / chunkSize

	// Initialize progress bar iterator
	tqdm.R(0, iterSize, func(v interface{}) (brk bool) {
		if index == iterSize {
			return false
		}

		currentChunk := readNBytes(serialPort, chunkSize, chunkSize*index)

		_, err = dumpFile.Write(currentChunk)
		if err != nil {
			log.Fatalln("File write error: ", err)
		}

		if cycleSleep != 0 {
			time.Sleep(time.Duration(cycleSleep) * time.Millisecond)
		}

		index++
		return
	})

	fmt.Println()
}
