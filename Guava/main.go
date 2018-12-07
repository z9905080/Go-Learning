package main

import (
	DurianApi "bbin/kernel/api"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"./api"
	"./struct"
	"github.com/BurntSushi/toml"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/subosito/gotenv"
)

var wsupgrader = websocket.Upgrader{}
var client_map = make(map[string]chan []byte)

//讀寫鎖
var mu sync.RWMutex

var conf Config

//config結構
type Config struct {
	Redis struct {
		Publish struct {
			IP      string `toml:"ip"`
			Host    string `toml:"host"`
			Port    string `toml:"port"`
			Channel string `toml:"channel"`
		} `toml:"publish"`
		Member struct {
			Master struct {
				IP   string `toml:"ip"`
				Host string `toml:"host"`
				Port string `toml:"port"`
			} `toml:"master"`
			Slave struct {
				IP   string `toml:"ip"`
				Host string `toml:"host"`
				Port string `toml:"port"`
			} `toml:"slave"`
		} `toml:"member"`
	} `toml:"redis"`
	Web struct {
		Port string `toml:"port"`
	} `toml:"web"`
	MaxLink struct {
		Max int64 `toml:"max"`
	} `toml:"maxlink"`
}

/*
	接收redis廣播
*/
func receiveRedis() {
	writelog("INFO", "receiveRedis Start!")

	redis_conn, err := redis.Dial("tcp", conf.Redis.Publish.IP+":"+conf.Redis.Publish.Port)
	if err != nil {
		logstr := fmt.Sprintf("redis conn err: %s\nData: IP %s Port %s", err, conf.Redis.Publish.IP, conf.Redis.Publish.Port)
		writelog("ERROR", logstr)
	}
	psc := redis.PubSubConn{redis_conn}
	psc.PSubscribe(conf.Redis.Publish.Channel)

	for {
		switch v := psc.Receive().(type) {
		case redis.PMessage:
			//json解碼
			var msg_list guava.MsgList
			err := json.Unmarshal(v.Data, &msg_list)
			if err != nil {
				logstr := fmt.Sprintf("json.Unmarshal err: %s\nData: %s\n", err, v.Data)
				writelog("ERROR", logstr)
			}

			for _, msg := range msg_list {
				go pushToUser(msg)
			}
		case redis.Subscription:
		}
	}
}

/*
	傳送訊息至對應user的ws_id的channel
*/
func pushToUser(msg_data guava.MsgData) {
	redis_conn, err := redis.Dial("tcp", conf.Redis.Member.Slave.IP+":"+conf.Redis.Member.Slave.Port)
	if err != nil {
		logstr := fmt.Sprintf("redis conn err: %s\nData: IP %s Port %s", err, conf.Redis.Member.Slave.IP, conf.Redis.Member.Slave.Port)
		writelog("ERROR", logstr)
	}
	defer func() {
		if p := recover(); p != nil {
			logstr := fmt.Sprintf("pushToUser panic: %v", p)
			writelog("ERROR", logstr)
		}
	}()

	//分析收件人
	for _, rec := range msg_data.Rec {
		//抓取該收件人（廳、層）在線會員
		online_users, err := redis.Values(redis_conn.Do("HKEYS", rec))
		if err != nil {
			logstr := fmt.Sprintf("redis err: %s\nData: HKEYS, %s\n", err, rec)
			writelog("ERROR", logstr)
		}
		for _, user := range online_users {
			//抓取該會員目前連線websocket id
			links, err := redis.Values(redis_conn.Do("LRANGE", string(user.([]byte))+"_ws", "0", "-1"))
			if err != nil {
				logstr := fmt.Sprintf("redis err: %s\n", err)
				writelog("ERROR", logstr)
			}
			for _, ws_id := range links {
				ws := string(ws_id.([]byte))
				//將訊息json編碼
				msg := guava.WsMsg{msg_data.CN.Subject, msg_data.EN.Subject, msg_data.TW.Subject, string(user.([]byte))}
				msg_json, err := json.Marshal(msg)
				if err != nil {
					logstr := fmt.Sprintf("json.Marshal err: %s\nData: %s", err, msg)
					writelog("ERROR", logstr)
				}
				//塞入channel
				mu.RLock() //讀取鎖
				if client_map[ws] != nil {
					client_map[ws] <- []byte(msg_json)
				}
				mu.RUnlock() //解讀取鎖
			}
		}
	}
	redis_conn.Close()
}

/*
	客端接入口
*/
func cliententer(c echo.Context) error {
	ws, err := wsupgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		logstr := fmt.Sprintf("upgrader.Upgrade err: %s\n", err)
		writelog("ERROR", logstr)
		return nil
	}

	_, client_data, err := ws.ReadMessage()
	if err != nil {
		logstr := fmt.Sprintf("ws.ReadMessage err: %s\n", err)
		writelog("ERROR", logstr)
		ws.Close()
		return nil
	}

	//json解碼SID與user id
	var user_info guava.UserInfo
	err = json.Unmarshal(client_data, &user_info)
	if err != nil {
		logstr := fmt.Sprintf("json.Unmarshal err: %s\nData: %v", err, client_data)
		writelog("ERROR", logstr)
	}

	//檢查SID & 紀錄上層IDß
	user_info, check := checkUserInfo(user_info)
	if check != true {
		//return errors.New("SID error!")
		return nil
	}

	//產生ws id
	user_id := user_info.MEM
	t := time.Now().UnixNano()
	ws_id := user_id + fmt.Sprintf("%d", t)

	//設定使用者資料到Redis
	setUserInfo(user_info, ws_id)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	mu.Lock() //獨佔鎖
	client_map[ws_id] = make(chan []byte, 3)
	mu.Unlock() //解獨佔鎖
	mu.RLock()  //讀取鎖
	go pushMsg(client_map[ws_id], ws, wg, ws_id, user_info)
	mu.RUnlock() //解讀取鎖
	wg.Wait()

	defer func() {
		if p := recover(); p != nil {
			logstr := fmt.Sprintf("cliententer panic: %v", p)
			writelog("ERROR", logstr)
		}
	}()
	mu.Lock() //獨佔鎖
	delete(client_map, ws_id)
	mu.Unlock() //解獨佔鎖
	return nil
}

/*
	檢查客端傳入的使用者資料
*/
func checkUserInfo(user_info guava.UserInfo) (guava.UserInfo, bool) {
	defer func() {
		if p := recover(); p != nil {
			logstr := fmt.Sprintf("checkUserInfo panic: %v", p)
			writelog("ERROR", logstr)
		}
	}()
	user_id := user_info.MEM

	//檢查sid正確性 & 設定使用者資料
	user_info, err := checkSID(user_info)
	if err != true {
		return user_info, false
	}

	//檢查連線數
	check_max_link := checkMaxLink(user_id)
	if check_max_link != true {
		return user_info, false
	}

	return user_info, true
}

/*
	檢查使用者session id
*/
func checkSID(user_info guava.UserInfo) (guava.UserInfo, bool) {
	sid := user_info.SID

	defer func() {
		if p := recover(); p != nil {
			logstr := fmt.Sprintf("checkSID panic: %v", p)
			writelog("ERROR", logstr)
		}
	}()

	//連線至炎五檢查sid
	rd5_check := API.GetSession(sid)
	if rd5_check.Result != "ok" {
		return user_info, false
	}

	//設定使用者上層id
	user_info.MEM = rd5_check.Ret.User.ID
	user_info.A = rd5_check.Ret.User.AllParents[0]
	user_info.SA = rd5_check.Ret.User.AllParents[1]
	user_info.C = rd5_check.Ret.User.AllParents[2]
	user_info.SC = rd5_check.Ret.User.AllParents[3]
	user_info.HALL = rd5_check.Ret.User.AllParents[4]

	return user_info, true
}

/*
	檢查客端連線是否已達最大連線數
	目前達到最大連線數時 會自動關閉最舊之連線
*/
func checkMaxLink(user_id string) bool {
	redis_conn, err := redis.Dial("tcp", conf.Redis.Member.Master.IP+":"+conf.Redis.Member.Master.Port)
	if err != nil {
		logstr := fmt.Sprintf("redis conn err: %s\nData: IP %s Port %s", err, conf.Redis.Member.Master.IP, conf.Redis.Member.Master.Port)
		writelog("ERROR", logstr)
	}
	redis_conn_s, err := redis.Dial("tcp", conf.Redis.Member.Slave.IP+":"+conf.Redis.Member.Slave.Port)
	if err != nil {
		logstr := fmt.Sprintf("redis conn err: %s\nData: IP %s Port %s", err, conf.Redis.Member.Slave.IP, conf.Redis.Member.Slave.Port)
		writelog("ERROR", logstr)
	}
	defer func() {
		if p := recover(); p != nil {
			logstr := fmt.Sprintf("checkMaxLink panic: %v", p)
			writelog("ERROR", logstr)
		}
		redis_conn.Close()
		redis_conn_s.Close()
	}()

	//目前連線數
	user_links, err := redis_conn_s.Do("LLEN", user_id+"_ws")
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: LLEN %s_ws", err, user_id)
		writelog("ERROR", logstr)
	}

	//超過上限時，刪除最舊的連線
	if conf.MaxLink.Max <= user_links.(int64) {
		old_links, err := redis_conn.Do("RPOP", user_id+"_ws")
		if err != nil {
			logstr := fmt.Sprintf("redis err: %s\nData: RPOP %s_ws", err, user_id)
			writelog("ERROR", logstr)
		}
		old_links_str := string(old_links.([]byte))
		//關閉舊連線的chan
		mu.RLock() //讀取鎖
		if client_map[old_links_str] != nil {
			mu.RUnlock() //解讀取鎖
			mu.Lock()    //獨佔鎖
			delete(client_map, old_links_str)
			mu.Unlock() //解獨佔鎖
		} else {
			mu.RUnlock() //解讀取鎖
		}
	}

	return true
}

/*
	將使用者資料寫進Redis
*/
func setUserInfo(user_info guava.UserInfo, ws_id string) {
	redis_conn, err := redis.Dial("tcp", conf.Redis.Member.Master.IP+":"+conf.Redis.Member.Master.Port)
	if err != nil {
		logstr := fmt.Sprintf("redis conn err: %s\nData: IP %s Port %s", err, conf.Redis.Member.Master.IP, conf.Redis.Member.Master.Port)
		writelog("ERROR", logstr)
	}
	defer func() {
		if p := recover(); p != nil {
			logstr := fmt.Sprintf("setUserInfo panic: %v", p)
			writelog("ERROR", logstr)
		}
	}()
	user_id := user_info.MEM

	//設定使用者ws_id
	_, err = redis_conn.Do("LPUSH", user_id+"_ws", ws_id)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: LPUSH %s_ws %s", err, user_id, ws_id)
		writelog("ERROR", logstr)
	}

	//設定redis到期時間（60*60*24）一天
	_, err = redis_conn.Do("EXPIRE", user_id+"_ws", "86400")
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: EXPIRE %s_ws 86400", err, user_id)
		writelog("ERROR", logstr)
	}

	//設定使用者資料
	_, err = redis_conn.Do("HSET", user_id, "HALL", user_info.HALL)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET %s HALL %s", err, user_id, user_info.HALL)
		writelog("ERROR", logstr)
	}
	_, err = redis_conn.Do("HSET", user_id, "SC", user_info.SC)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET %s SC %s", err, user_id, user_info.SC)
		writelog("ERROR", logstr)
	}

	_, err = redis_conn.Do("HSET", user_id, "C", user_info.C)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET %s C %s", err, user_id, user_info.C)
		writelog("ERROR", logstr)
	}

	_, err = redis_conn.Do("HSET", user_id, "SA", user_info.SA)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET %s SA %s", err, user_id, user_info.SA)
		writelog("ERROR", logstr)
	}

	_, err = redis_conn.Do("HSET", user_id, "A", user_info.A)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET %s A %s", err, user_id, user_info.A)
		writelog("ERROR", logstr)
	}

	_, err = redis_conn.Do("HSET", user_id, "LV", user_info.LV)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET %s LV %s", err, user_id, user_info.LV)
		writelog("ERROR", logstr)
	}

	//設定redis到期時間（60*60*24）一天
	_, err = redis_conn.Do("EXPIRE", user_id, "86400")
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: EXPIRE %s 86400", err, user_id)
		writelog("ERROR", logstr)
	}

	//將使用者寫入上層
	_, err = redis_conn.Do("HSET", "HALL_"+user_info.HALL, user_id, user_id)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET HALL_%s  %s %s", err, user_info.HALL, user_id, user_id)
		writelog("ERROR", logstr)
	}
	_, err = redis_conn.Do("HSET", "SC_"+user_info.SC, user_id, user_id)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET SC_%s  %s %s", err, user_info.SC, user_id, user_id)
		writelog("ERROR", logstr)
	}
	_, err = redis_conn.Do("HSET", "C_"+user_info.C, user_id, user_id)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET C_%s  %s %s", err, user_info.C, user_id, user_id)
		writelog("ERROR", logstr)
	}
	_, err = redis_conn.Do("HSET", "SA_"+user_info.SA, user_id, user_id)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET SA_%s  %s %s", err, user_info.SA, user_id, user_id)
		writelog("ERROR", logstr)
	}
	_, err = redis_conn.Do("HSET", "A_"+user_info.A, user_id, user_id)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET A_%s  %s %s", err, user_info.A, user_id, user_id)
		writelog("ERROR", logstr)
	}
	_, err = redis_conn.Do("HSET", "LV_"+user_info.LV, user_id, user_id)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET LV_%s  %s %s", err, user_info.LV, user_id, user_id)
		writelog("ERROR", logstr)
	}
	_, err = redis_conn.Do("HSET", "MEM_"+user_id, user_id, user_id)
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: HSET MEM_%s  %s %s", err, user_id, user_id, user_id)
		writelog("ERROR", logstr)
	}
	//設定redis到期時間（60*60*24）一天
	_, err = redis_conn.Do("EXPIRE", "MEM_"+user_id, "86400")
	if err != nil {
		logstr := fmt.Sprintf("redis err: %s\nData: EXPIRE MEM_%s 86400", err, user_id)
		writelog("ERROR", logstr)
	}
	redis_conn.Close()
}

/*
	推送ws訊息
*/
func pushMsg(cs chan []byte, ws *websocket.Conn, wg *sync.WaitGroup, ws_id string, user_info guava.UserInfo) {
	defer func() {
		if p := recover(); p != nil {
			logstr := fmt.Sprintf("pushMsg panic: %v", p)
			writelog("ERROR", logstr)
		}
	}()

	//偵測客端是否斷線
	go readLoop(cs, ws, ws_id, user_info)

	//等待chan，把訊息傳送至客端
	for {
		data := <-cs
		err := ws.WriteMessage(websocket.TextMessage, []byte(data))
		if err != nil {
			if data != nil {
				logstr := fmt.Sprintf("WriteMessage err: %s\nData: %s", err, data)
				writelog("ERROR", logstr)
			}
			wg.Done()
			break
		}
	}
}

/*
	偵測客端連線
*/
func readLoop(cs chan []byte, ws *websocket.Conn, ws_id string, user_info guava.UserInfo) {
	//持續偵測連線是否存在
	for {
		if _, _, err := ws.NextReader(); err != nil {
			//關閉websocket
			ws.Close()
			//關閉chan
			close(cs)
			mu.Lock() //獨佔鎖
			delete(client_map, ws_id)
			mu.Unlock() //解獨佔鎖

			redis_conn, err := redis.Dial("tcp", conf.Redis.Member.Master.IP+":"+conf.Redis.Member.Master.Port)
			if err != nil {
				logstr := fmt.Sprintf("redis conn err: %s\nData: IP %s Port %s", err, conf.Redis.Member.Master.IP, conf.Redis.Member.Master.Port)
				writelog("ERROR", logstr)
			}
			defer func() {
				if p := recover(); p != nil {
					logstr := fmt.Sprintf("readLoop panic: %v", p)
					writelog("ERROR", logstr)
				}
			}()
			//刪除redis 使用者的 ws_id
			_, err = redis_conn.Do("LREM", user_info.MEM+"_ws", "0", ws_id)
			if err != nil {
				logstr := fmt.Sprintf("redis err: %s\nData: LREM %s_ws 0", err, user_info.MEM)
				writelog("ERROR", logstr)
			}

			//檢查此使用者是否有剩餘的連線
			//目前連線數
			user_links, err := redis_conn.Do("LLEN", user_info.MEM+"_ws")
			if err != nil {
				logstr := fmt.Sprintf("redis err: %s\nData: LLEN %s_ws", err, user_info.MEM)
				writelog("ERROR", logstr)
			}

			//都無連線時 刪除user id
			if user_links == nil || user_links.(int64) < 1 {
				_, err = redis_conn.Do("HDEL", "HALL_"+user_info.HALL, user_info.MEM)
				if err != nil {
					logstr := fmt.Sprintf("redis err: %s\nData: HDEL HALL_%s %s\n", err, user_info.HALL, user_info.MEM)
					writelog("ERROR", logstr)
				}
				_, err = redis_conn.Do("HDEL", "SC_"+user_info.SC, user_info.MEM)
				if err != nil {
					logstr := fmt.Sprintf("redis err: %s\nData: HDEL SC_%s %s\n", err, user_info.SC, user_info.MEM)
					writelog("ERROR", logstr)
				}
				_, err = redis_conn.Do("HDEL", "C_"+user_info.C, user_info.MEM)
				if err != nil {
					logstr := fmt.Sprintf("redis err: %s\nData: HDEL C_%s %s\n", err, user_info.C, user_info.MEM)
					writelog("ERROR", logstr)
				}
				_, err = redis_conn.Do("HDEL", "SA_"+user_info.SA, user_info.MEM)
				if err != nil {
					logstr := fmt.Sprintf("redis err: %s\nData: HDEL SA_%s %s\n", err, user_info.SA, user_info.MEM)
					writelog("ERROR", logstr)
				}
				_, err = redis_conn.Do("HDEL", "A_"+user_info.A, user_info.MEM)
				if err != nil {
					logstr := fmt.Sprintf("redis err: %s\nData: HDEL A_%s %s\n", err, user_info.SA, user_info.MEM)
					writelog("ERROR", logstr)
				}
				_, err = redis_conn.Do("HDEL", "LV_"+user_info.LV, user_info.MEM)
				if err != nil {
					logstr := fmt.Sprintf("redis err: %s\nData: HDEL LV_%s %s\n", err, user_info.LV, user_info.MEM)
					writelog("ERROR", logstr)
				}
				_, err = redis_conn.Do("HDEL", "MEM_"+user_info.MEM, user_info.MEM)
				if err != nil {
					logstr := fmt.Sprintf("redis err: %s\nData: HDEL MEM_%s %s\n", err, user_info.MEM, user_info.MEM)
					writelog("ERROR", logstr)
				}
			}
			redis_conn.Close()
			break
		}
	}
}

/*
	publish test
*/
/*
func publishsend(c echo.Context) error {
	name := c.FormValue("name")
	passwd := c.FormValue("passwd")

	if name != "ragnarok-ctl_a" || passwd != "2tE5j8V7Ux7G"{
		return errors.New("login error!")
	}

	rec := c.FormValue("rec")

	redis_conn, err := redis.Dial("tcp", conf.Redis.Publish.IP + ":" + conf.Redis.Publish.Port)
	if err != nil {
		logstr := fmt.Sprintf("redis conn err: %s\nData: IP %s Port %s", err, conf.Redis.Publish.IP, conf.Redis.Publish.Port)
		writelog("ERROR", logstr)
	}
	defer func() {
		if p := recover(); p != nil {
			logstr := fmt.Sprintf("publishsend panic: %v", p)
			writelog("ERROR", logstr)
		}
	}()

	data := `[{"CN":{"subject":"測試"},"EN":{"subject":"測試"},"TW":{"subject":"測試"},"rec":["MEM_` + rec + `"],"msgid":999}]`
	redis_conn.Do("PUBLISH", "wsmsg" , data)
	redis_conn.Close()

	return c.String(http.StatusOK, "Redis廣播發送成功")
}
*/

func writelog(tag string, msg string) {
	//設定時間
	now := time.Now().Format("2006-01-02 15:04:05")
	year := time.Now().Format("2006")
	month := time.Now().Format("1")
	day := time.Now().Format("2")
	fileName := "server.log"

	//檢查今日log檔案是否存在
	if _, err := os.Stat("logs/" + year + "-" + month + "-" + day + "/" + fileName); os.IsNotExist(err) {
		//建立資料夾
		folderPath := os.Getenv("ROOT") + "/logs/" + year + "-" + month + "-" + day
		os.MkdirAll(folderPath, 0777)
		//建立檔案
		_, err := os.Create("logs/" + year + "-" + month + "-" + day + "/" + fileName)
		if err != nil {
			fmt.Println("open file error !")
			fmt.Println(err)
		}
	}

	//開啟檔案準備寫入
	logFile, err := os.OpenFile("logs/"+year+"-"+month+"-"+day+"/"+fileName, os.O_RDWR|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println("open file error !")
		fmt.Println(err)
	}

	logFile.WriteString(now + " [" + tag + "] " + msg + "\n")
	logFile.Close()
}

func main() {
	//載入config
	gotenv.Load()
	cf := os.Getenv("ROOT") + "config/" + os.Getenv("ENV") + "_config.toml"
	if _, err := toml.DecodeFile(cf, &conf); err != nil {
		fmt.Println(err)
	}

	//API連線設定
	DurianApi.Init(cf)

	writelog("INFO", "Server Start!")
	go receiveRedis()

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//TEST
	//e.Static("/guava/test", "public/websocket.html")
	//e.Static("/guava/publish", "public/publish.html")
	//e.POST("/guava/publishsend", publishsend)

	e.GET("/guava/ws", cliententer)
	e.Logger.Fatal(e.Start(":" + conf.Web.Port))

	defer func() {
		writelog("INFO", "Server down!")
	}()
}
