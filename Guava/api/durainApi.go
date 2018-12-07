package API
import (
	"os"
	"fmt"
	"log"
	"time"
	"encoding/json"
	"github.com/subosito/gotenv"
	"../struct"
	"bbin/kernel/api"
)

var serverLog *log.Logger

func init() {
	gotenv.Load()

	//set Log
	now := time.Now().Format("20060102150405")
	folderPath := os.Getenv("ROOT") + "/logs/api"
	fileName := now + ".log"
	os.MkdirAll(folderPath, 0777)
	logFile, err  := os.Create("logs/api/" + fileName)
	defer logFile.Close()
	if err != nil {
		fmt.Println("open file error !")
		fmt.Println(err)
	}
	serverLog = log.New(logFile,"[Error]",log.LstdFlags)
}

// GET /api/session/{sessionId}
func GetSession(sid string) guava.SessionData {
	url := "session/" + sid
	p := API.Param {}

	var result guava.SessionData
	data := API.Get(url, p)
	err := json.Unmarshal(data, &result)
	if err != nil {
		serverLog.SetPrefix("[ERROR]")
		serverLog.Println("Json Decode Error", err)
	}

	return result
}