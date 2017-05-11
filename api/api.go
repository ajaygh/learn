/*This is a middleware code for recieving middleware data from
middleware server and send it to api-server and vice-versa
Author: gor
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	PORT          = 5555
	HOST          = "127.0.0.1"
	SIZE          = 512
	CONF_FILE     = "api"
	CONF_PATH     = "."
	MAX_DATA_SIZE = 512
	API_ID        = 0x09
	API_NUM       = 0x01
	MIDDLEWARE_ID = 0x02
	DSI           = 17
)

//type for different kinds of recievid data
type RcvType uint8

const (
	SCAN RcvType = iota + 1
	SORT_SORTER
	SORT_ECDS
	EVENT
	INIT_CONFIG
	SORT_ICR
)

var Conf map[string]string

func main() {
	readConfig(CONF_FILE, CONF_PATH, "production")
	RcvFrom(HOST, PORT)
}

func RcvFrom(host string, port int) {
	addr := host + ":" + strconv.Itoa(port)
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)
	checkError("panic", "resolve udp failed", err)

	udpConn, err := net.ListenUDP("udp", udpAddr)
	checkError("panic", "listen failed", err)

	//Listen forever
	fmt.Println("Started Communicating with middleware server on port", port)
	for {
		handleConn(udpConn)
	}
}

func checkApiData(buf []byte) {
	//check if given data is for api
	if buf[5] != API_NUM && buf[6] != API_ID {
		log.Fatalln("Receiver mismatch.")
		return
	}
	//check if received packets can be processed or not
	if buf[16] != MIDDLEWARE_ID {
		log.Fatalln("Packet type mismatch")
		return
	}
}

func handleConn(conn *net.UDPConn) {
	//	defer conn.Close()

	buf := make([]byte, SIZE)
	_, addr, err := conn.ReadFromUDP(buf[0:])
	checkError("panic", "Error in data", err)

	//get type and handle accordingly
	switch rcvType := uint8(buf[15]); RcvType(rcvType) {
	case SCAN:
		//go processScanData(buf)
	case SORT_SORTER, SORT_ECDS, SORT_ICR:
		go processSortData(buf)
	case EVENT:
		//processEventData(buf)
	default:
		fmt.Fprintf(os.Stderr, "Wrong data type received.\n")
	}
	conn.WriteToUDP([]byte("check handleConn"), addr)
}

func processScanData(data []byte) {
	//filter all relevant data and put them into mapScan

	//packetLen := int(data[9]) + int(data[10])*256
	icrID := strconv.Itoa(int(data[DSI]))
	jobID := strconv.Itoa(int(data[DSI+1]) + int(data[DSI+2])*256 +
		int(data[DSI+3])*65536 + int(data[DSI+4])*16777216)
	casketID := strconv.Itoa(int(data[DSI+5]) + int(data[DSI+6])*256)
	width := strconv.Itoa(int(data[DSI+7]) + int(data[DSI+8])*256)
	length := strconv.Itoa(int(data[DSI+9]) + int(data[DSI+10])*256)
	height := strconv.Itoa(int(data[DSI+11]) + int(data[DSI+12])*256)
	boxVol := strconv.Itoa(int(data[DSI+13]) + int(data[DSI+14])*256 +
		int(data[DSI+15])*65536)
	realVol := strconv.Itoa(int(data[DSI+16]) + int(data[DSI+17])*256 +
		int(data[DSI+18])*65536)
	volStatus := strconv.Itoa(int(data[DSI+19]) + int(data[DSI+20])*256)
	weightStatus := strconv.Itoa(int(data[DSI+21]))
	weight := strconv.Itoa(int(data[DSI+22]) + int(data[DSI+23])*256)
	inputNo := strconv.Itoa(int(data[DSI+24]))

	imageDay := data[DSI+25]
	imageMonth := data[DSI+26]
	imageYear := int(data[DSI+27]) + int(data[DSI+28])*256
	imageHour := data[DSI+29]
	imageMinutes := data[DSI+30]
	imageSeconds := data[DSI+31]
	imageMilliseconds := int(data[DSI+32]) + int(data[DSI+33])*256
	imageUniqueNumber := int(data[DSI+34]) + int(data[DSI+35])*256

	imageID := fmt.Sprintf("%04d-%02d-%02d-%04d-%02d-%02d-%02d-%03d",
		imageUniqueNumber, imageDay, imageMonth, imageYear,
		imageHour, imageMinutes, imageSeconds, imageMilliseconds)

	//strncpy(uuid, barVms.uuid, UUID_LENGTH)
	uuid := "1234567891234567891234567891234567890"
	//strcpy(barcode, barVms.barcode)
	barcode := "aa_bb_cc\ndd__ee"
	numOfBarcodes := data[DSI+36]
	scanStatus := Conf["scan_success"]
	if numOfBarcodes == 0 {
		scanStatus = Conf["scan_failure"]
	}

	scan := &Scan{icrID, jobID, casketID, width, length, height,
		boxVol, realVol, volStatus, weightStatus, weight, inputNo,
		imageID, uuid, barcode, scanStatus}

	chuteId := sendScanData(scan)
	fmt.Printf("CHUTEID RECEIVED = %v\n", chuteId)
	//makeChuteIdPacket(apiChuteSnd, jobId, casketId, chuteId);
	//mcastApiSend(chutePacket);
}

func processSortData(data []byte) {
	jobID := strconv.Itoa(int(data[DSI+1]) + int(data[DSI+2])*256 +
		int(data[DSI+3])*65536 + int(data[DSI+4])*16777216)
	casketID := strconv.Itoa(int(data[DSI+5]) + int(data[DSI+6])*256)
	chuteID := strconv.Itoa(int(data[DSI+7]) + int(data[DSI+8])*256)
	sortStatus := strconv.Itoa(int(data[DSI+9]))
	//uuid 10 to 46
	uuid := "1234567891234567891234567891234567890"
	sorterID := strconv.Itoa(int(data[DSI+47]))
	sort := &Sort{jobID, uuid, sortStatus, chuteID}

	fmt.Printf(`SORT RECEIVED : JOB ID - %s CASKET_ID - %s CHUTE_ID - %s UUID - %s,
	SORT STATUS - %s\n`, jobID, casketID, chuteID, uuid, sortStatus)

	switch sorterType := uint8(data[15]); RcvType(sorterType) {
	case SORT_SORTER:
		sendSortData(sort, SORT_SORTER, sorterID)
	case SORT_ECDS:
		sendSortData(sort, SORT_ECDS, sorterID)
	case SORT_ICR:
		sendSortData(sort, SORT_ICR, sorterID)
	}
}

func processEventData(data []byte) {
	//request to corresponding url
	//sendEventData(data)
}

func sendScanData(scan *Scan) int {
	req := makeRequest(scan, "POST", Conf["scan_url"])
	client := &http.Client{Timeout: time.Second}
	resp, err := client.Do(req)
	checkError("fatal", "req execution failed", err)

	defer resp.Body.Close()
	decoder := json.NewDecoder(resp.Body)
	scanRcv := ScanRcv{}
	err = decoder.Decode(&scanRcv)
	checkError("fatal", "decoding failed", err)

	log.Println("sendSD : chuteid = ", scanRcv.ChuteId)
	//send chute id
	chuteID, _ := strconv.Atoi(scanRcv.ChuteId)

	return chuteID
}

func sendSortData(data *Sort, sortType RcvType, sorterID string) {
	var req *http.Request
	switch sortType {
	case SORT_SORTER:
		sorterData := &SorterSort{*data, sorterID}
		req = makeRequest(sorterData, "POST", Conf["sort_url"])
	case SORT_ECDS:
		ecdsSortData := &EcdsSort{*data, sorterID}
		req = makeRequest(ecdsSortData, "POST", Conf["feedback_url"])
	case SORT_ICR:
		sorterData := &IcrSort{*data, sorterID}
		req = makeRequest(sorterData, "POST", Conf["sort_url"])
	}

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Do(req)
	checkError("fatal", "req execution failed", err)

	defer resp.Body.Close()
	var res interface{}
	json.NewDecoder(resp.Body).Decode(&res)
	fmt.Println("Sort response received ", res)
}

func sendEventData(data *Event) {
	//prepare and send event
	//make request handle
	// send sequentially
}
