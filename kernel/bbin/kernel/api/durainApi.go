package API

import (
	"io"
	"fmt"
	"time"
	"strings"
	"io/ioutil"
	"net/http"
	"net/url"
	"github.com/BurntSushi/toml"
	"bbin/kernel/log"
	"encoding/json"
)

var conf Config
var Client http.Client

type Param map[string][]string
type Config struct {
	API struct {
		    Durian struct {
				   IP   string `toml:"ip"`
				   Host string `toml:"host"`
				   Port string `toml:"port"`
			   } `toml:"durian"`
		    Ipl    struct {
				   IP   string `toml:"ip"`
				   Host string `toml:"host"`
				   Port string `toml:"port"`
			   } `toml:"ipl"`
	    } `toml:"api"`
}

func Init(path string)  {
	if _, err := toml.DecodeFile(path, &conf); err != nil {
		// handle error
		fmt.Println(err)
	}

	Client = http.Client{
		Timeout: time.Second * 300,
	}

	fmt.Printf("Load Api config success.\n Api: %v\n\n", conf)
}

func Get(apiPath string, p Param) []byte {
	return connect("GET", apiPath, p)
}

func Put(apiPath string, p Param) []byte {
	return connect("PUT", apiPath, p)
}

func Post(apiPath string, p Param) []byte {
	return connect("POST", apiPath, p)
}

func Delete(apiPath string, p Param) []byte {
	return connect("DELETE", apiPath, p)
}

func connect(method string, apiPath string, p Param) []byte {
	ip := conf.API.Durian.IP
	hostName := conf.API.Durian.Host
	new_url := "http://" + ip + "/api/" + apiPath

	var rData string
	for key, value := range p {
		for _, v := range value {
			rData += key + "=" + url.QueryEscape(v) + "&"
		}
	}

	// Set request data
	var payload io.Reader
	switch method {
	case "GET":
		new_url += "?" + rData
		payload = nil
	case "POST":
		payload = strings.NewReader(rData)
	case "PUT":
		payload = strings.NewReader(rData)
	case "DELETE":
		payload = strings.NewReader(rData)
	}

	req, err := http.NewRequest(method, new_url, payload); if err != nil {
		return errhandler(err)
	}
	// HEADER 底加拉
	req.Host = hostName
	if method != "GET" {
		req.Header.Add("content-type", "application/x-www-form-urlencoded")
	}

	res, getErr := Client.Do(req); if getErr != nil {
		return errhandler(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body); if readErr != nil {
		return errhandler(readErr)
	}
	defer res.Body.Close()

	return body
}

func connectLog(logData error)  {
	now := time.Now().UTC().Add(time.Duration(-4) * time.Hour)

	year := now.Format("2006")
	month := now.Format("01")
	day := now.Format("02")
	path := "logs/" + year + "-" + month + "-" + day + "/connect_error.log"

	Log.LogFile(path, logData.Error())
}

func errhandler(err error) []byte {
	fmt.Println(err)
	connectLog(err)

	errStr, _ := json.Marshal(err)
	return errStr
}
