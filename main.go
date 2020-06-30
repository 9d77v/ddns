package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
)

var (
	accessKeyID = os.Getenv("AccessKeyID")
	secret      = os.Getenv("SECRET")
	domainName  = os.Getenv("DomainName")
	rr          = os.Getenv("RR")
)

func main() {
	client := getClient()
	currentIP, recordID := getCurrentIP(client)
	log.Println("current ip:", currentIP)
	log.Println("record id:", recordID)

	publicIP := getPublicIP()
	log.Println("public ip:", publicIP)
	if publicIP != currentIP {
		updateIP(client, publicIP, recordID)
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

func getCurrentIP(client *alidns.Client) (string, string) {
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.Scheme = "https"
	request.DomainName = domainName
	request.RRKeyWord = rr

	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		log.Panicln("get domain record failed", err)
	}
	records := response.DomainRecords.Record
	if len(records) == 0 {
		log.Panicln("dns records was empty")
	}
	return records[0].Value, records[0].RecordId
}

func updateIP(client *alidns.Client, recordID, publicIP string) {
	request := alidns.CreateUpdateDomainRecordRequest()
	request.Scheme = "https"

	request.RecordId = recordID
	request.RR = rr
	request.Type = "A"
	request.Value = publicIP

	response, err := client.UpdateDomainRecord(request)
	if err != nil {
		log.Println("update ip faield", err.Error())
	}
	log.Printf("update ip success,response is %#v\n", response)
}

func getPublicIP() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		log.Panicln("get public ip failed:", err)
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	return string(content)
}
