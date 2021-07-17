package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

var (
	accessKeyID = os.Getenv("AccessKeyID")
	secret      = os.Getenv("SECRET")
	domainName  = os.Getenv("DomainName")
	rr          = os.Getenv("RR")
	ipDomain    = os.Getenv("GET_IP_DOMAIN")
)

func main() {
	rrSet := map[string]struct{}{}
	for _, v := range strings.Split(rr, ",") {
		rrSet[v] = struct{}{}
	}
	client := getClient()
	records := getCurrentIP(client)
	validRecords := []alidns.Record{}
	for _, v := range records {
		if _, ok := rrSet[v.RR]; ok {
			log.Printf("domain: %s current ip: %s record id: %s", v.RR, v.Value, v.RecordId)
			validRecords = append(validRecords, v)
		}
	}
	publicIP := getPublicIP()
	log.Println("public ip:", publicIP)
	if publicIP != validRecords[0].Value {
		updateIP(client, validRecords, publicIP)
	} else {
		log.Println("ip not change")
	}
}

func getClient() *alidns.Client {
	client, err := alidns.NewClientWithAccessKey(
		"cn-hangzhou",
		accessKeyID,
		secret)
	if err != nil {
		log.Panicln("create ecs client failed")
	}
	return client
}

func getCurrentIP(client *alidns.Client) []alidns.Record {
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = domainName

	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		log.Panicln("get domain record failed", err)
	}
	records := response.DomainRecords.Record
	if len(records) == 0 {
		log.Panicln("dns records was empty")
	}
	return records
}

func updateIP(client *alidns.Client, records []alidns.Record, publicIP string) {
	for _, record := range records {
		request := alidns.CreateUpdateDomainRecordRequest()
		request.Scheme = "https"
		request.RecordId = record.RecordId
		request.RR = record.RR
		request.Type = "A"
		request.Value = publicIP
		response, err := client.UpdateDomainRecord(request)
		if err != nil {
			log.Println("update ip faield", err.Error())
		}
		log.Printf("update ip success,response is %#v\n", response)
	}
}

func getPublicIP() string {
	resp, err := http.Get(ipDomain)
	if err != nil {
		log.Panicln("get public ip failed:", err)
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return string(content)
}
