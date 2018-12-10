package guavaStruct

type SessionData struct {
	Result string `json:"result"`
	Ret    struct {
		Session struct {
			ID         string      `json:"id"`
			CreatedAt  string      `json:"created_at"`
			ModifiedAt interface{} `json:"modified_at"`
		} `json:"session"`
		User struct {
			ID          string   `json:"id"`
			Username    string   `json:"username"`
			Domain      string   `json:"domain"`
			Alias       string   `json:"alias"`
			Sub         bool     `json:"sub"`
			Test        bool     `json:"test"`
			HiddenTest  bool     `json:"hidden_test"`
			Size        string   `json:"size"`
			Currency    string   `json:"currency"`
			CreatedAt   string   `json:"created_at"`
			ModifiedAt  string   `json:"modified_at"`
			LastLogin   string   `json:"last_login"`
			Role        string   `json:"role"`
			AllParents  []string `json:"all_parents"`
			Ingress     string   `json:"ingress"`
			ClientOs    string   `json:"client_os"`
			LastLoginIP string   `json:"last_login_ip"`
		} `json:"user"`
		Cash struct {
			Currency string `json:"currency"`
		} `json:"cash"`
		IsMaintaining struct {
		} `json:"is_maintaining"`
		Whitelist []interface{} `json:"whitelist"`
	} `json:"ret"`
	Profile struct {
		ExecutionTime string `json:"execution_time"`
		ServerName    string `json:"server_name"`
	} `json:"profile"`
}

type UserInfo struct {
	MEM  string `json:"MEM"`
	SID  string `json:"sid"`
	HALL string `json:"Hall"`
	SC   string `json:"SC"`
	C    string `json:"C"`
	SA   string `json:"SA"`
	A    string `json:"A"`
	LV   string `json:"LV"`
}

type MsgList []struct {
	CN struct {
		Subject string `json:"subject"`
	} `json:"CN"`
	EN struct {
		Subject string `json:"subject"`
	} `json:"EN"`
	TW struct {
		Subject string `json:"subject"`
	} `json:"TW"`
	Rec   []string `json:"rec"`
	Msgid int      `json:"msgid"`
}

type MsgData struct {
	CN struct {
		Subject string `json:"subject"`
	} `json:"CN"`
	EN struct {
		Subject string `json:"subject"`
	} `json:"EN"`
	TW struct {
		Subject string `json:"subject"`
	} `json:"TW"`
	Rec   []string `json:"rec"`
	Msgid int      `json:"msgid"`
}

type WsMsg struct {
	CN   string `json:"zh-cn"`
	EN   string `json:"en"`
	TW   string `json:"zh-tw"`
	USER string `json:"user"`
}
