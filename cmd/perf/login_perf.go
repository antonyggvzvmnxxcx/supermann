package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/anyaddres/supermann/api"

	"github.com/icrowley/fake"
)

const (
	// Number of iterations
	Num = 3000000
	// Url to call
	Url = "http://127.0.0.1:8080/api/identifylogins/"
)

func randomTS() int64 {
	randomTime := rand.Int63n(time.Now().Unix()-94608000) + 94608000
	return time.Unix(randomTime, 0).Unix()
}

func randomIP() string {
	return fake.IPv4()
}

func main() {

	for index := 0; index < Num; index++ {
		event := api.LoginRequest{UserName: "bob", UnixTimeStamp: randomTS(), IpAddress: randomIP(),
			EventUUID: "85ad929a-db03-4bf4-9541-8f728fa12e42"}
		jsonStr, err := json.Marshal(event)
		if err != nil {
			panic(err)
		}
		req, err := http.NewRequest("POST", Url, bytes.NewBuffer(jsonStr))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		t1 := time.Now()
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		log.Println("The request took", time.Since(t1))
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println(string(body))
	}
}
