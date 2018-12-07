package Log

import (
	"bbin/kernel/basic"
	"fmt"
	"log"
	"os"
	"strings"
)

// Log
func LogFile(path string, logString string) {
	// 擷取字串檔案名稱與資料夾路徑
	s := strings.Split(path, "/")
	fileName := s[len(s)-1]
	filePath := strings.Split(path, fileName)[0]

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.MkdirAll(filePath, 0777)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	log.SetOutput(f)
	defer log.SetOutput(os.Stderr)

	log.Println(logString)
}

// DeleteOldLog 刪除舊的log
func DeleteOldLog(dirPath string, keepDay int) {
	logPath := ""
	dayList := Basic.GetDayBeforeMonth("US_EST", keepDay)

	for i := 0; i < 30; i++ {
		logPath = dirPath + "/" + dayList[i]

		// 確認檔案
		if _, err := os.Stat(logPath); os.IsNotExist(err) {
			// 沒找到檔案就不動作
		} else {
			removeErr := os.RemoveAll(logPath)
			if removeErr != nil {
				fmt.Println(removeErr.Error())
			}

			fmt.Println("Log already remove: " + logPath)
		}
	}
}
