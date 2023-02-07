package main

// The x8 Nano modbus data collection
// Data: 32bit float,
// Date:20230206-0207

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/goburrow/modbus"
)

const (
	RegQuantity        uint16 = 96     // the driver supports up to 125, need divide into two parts 96,96->192
	StartNanoSnAddress uint16 = 0xAFC8 //45000
	// StartNanoSnAddress uint16 = 0xAFCA //45002

	MaxNanoNum      = 16
	NanoSnByte      = 8 // for each nano sn
	DefaultHost     = "192.168.179.5"
	DefaultPort     = "502" //standard
	DefaultRate     = 60    //second
	MinRate         = 10
	DefaultDataNums = 100 //
	MaxDataNums     = 1000
)

type nanoInfo struct {
	SN           string
	startRegAddr uint16
}

var x8NanoList [16]nanoInfo

func main() {

	// get device host (url or ip address) and port from the command line
	var (
		host     string
		port     string
		rate     int64
		dataNums int64
	)

	flag.StringVar(&host, "host", DefaultHost, "Slave device host (url or ip address)")
	flag.StringVar(&port, "port", DefaultPort, fmt.Sprintf("Slave device port (default:%s)", DefaultPort))
	flag.Int64Var(&rate, "rate", DefaultRate, "Data collection rate in Second. > 10 required.")
	flag.Int64Var(&dataNums, "nums", DefaultDataNums, fmt.Sprintf("Number (Default:%d Max:%d) of data to collect.", DefaultDataNums, MaxDataNums))

	flag.Parse()
	if rate < MinRate {
		rate = MinRate
	}
	if dataNums > MaxDataNums {
		dataNums = MaxDataNums
	}
	mbHandler := modbus.NewTCPClientHandler(host + ":" + port)
	mbHandler.Timeout = 10 * time.Second
	mbHandler.SlaveId = 1

	var err error

	if err = mbHandler.Connect(); err != nil {
		log.Fatal("Connect error:", err)
	}
	defer mbHandler.Close()

	client := modbus.NewClient(mbHandler)
	printNanoDataHeader()

	for i := 0; i < int(dataNums); i++ {
		readNanoSn(client)
		fmt.Println()
		time.Sleep(time.Duration(rate) * time.Second)
	}
}

// readNanoSn: To readin all connected nano sn upto 16.
func readNanoSn(client modbus.Client) {
	x8Data, err := client.ReadHoldingRegisters(StartNanoSnAddress, MaxNanoNum*6) // manual wrote 5:NG

	if err != nil {
		fmt.Println("Read holding reg error.", err)
	}

	for i := 0; i < len(x8Data); i++ {
		fmt.Printf("%v,", x8Data[i])
	}
	fmt.Println()
	for i := 0; i < MaxNanoNum; i++ {
		fmt.Printf("Nano No.%d..........\n", i+1)
		for j := 0; j < 4; j++ {
			fmt.Printf("%v,%v..", x8Data[j*2+i*12], x8Data[j*2+i*12+1])
		}
		sn := string(x8Data[i*12 : i*12+8])
		fmt.Printf("SN:=%s..", sn)
		ad := uint16(x8Data[10+i*12]) | uint16(x8Data[11+i*12])<<8
		fmt.Printf("Reg:(%d, %X)\n", ad, ad)

		if ad < 45035 && ad > 45001 && strings.Contains(sn, "Z2R1") {
			x8NanoList[i].SN = sn
			x8NanoList[i].startRegAddr = ad
		} else {
			x8NanoList[i].SN = ""
			x8NanoList[i].startRegAddr = 0
			fmt.Println("Cannot reg:", x8NanoList[i])
		}
	}

	for i := 0; i < MaxNanoNum; i++ {
		if x8NanoList[i].startRegAddr > 0 {
			fmt.Printf("No.%d %v\n", i, x8NanoList[i])
		}
	}
}

// func getNanoData(client modbus.Client, startAddr uint16) {

func printNanoDataHeader() {
	fmt.Printf("ZARK Nano Modbus Data\n----------------------\n")
	fmt.Print("Time, ")
	fmt.Printf("X, Y, Z axis, CIV\n")
}
