package main

import (
	"strconv"
	"strings"
)

func main() {
	GetMachineGame("2")
}

// GetMachineGame 取得機率遊戲
func GetMachineGame(MID string) (ErrorCode, error) {

	data := getMachineGame(MID)
	if len(data) == 0 {
		return DataEmpty, nil
	}

	var totalBalanceInGame float64

	for _, item := range data {
		s := strings.Split(item.BetBase, ":")
		// 用冒號切開後有兩個(才有要處理的資料)
		if len(s) >= 2 {
			child, err1 := strconv.Atoi(s[0])
			parent, err2 := strconv.Atoi(s[1])

			// 都沒轉換錯誤才處理
			if err1 == nil && err2 == nil {
				if parent != 0 {
					totalBalanceInGame += item.Balance / float64(parent) * float64(child)
				}
			} else if err1 != nil {
				return BetBaseWrong, err1
			} else {
				return BetBaseWrong, err2
			}

		} else {
			return BetBaseEmpty, nil
		}
	}
	if totalBalanceInGame > 0 {
		return OK, nil
	}

	return DataEmpty, nil
}

// getMachineGame 取機台機率的餘額
// case1 : 可能select 不到資料
// case2 : BetBase 的內容可能是 1:1000, 100:1, 1:1, 等開分的比例
// case3 : BetBase可能為空字串
func getMachineGame(UserID string) []MachineInfo {

	machineInfo := []MachineInfo{
		MachineInfo{
			BetBase: "100:1",
			Balance: 123,
			UserID:  1,
		},
		MachineInfo{
			BetBase: "100:1",
			Balance: 123,
			UserID:  2,
		},
		MachineInfo{
			BetBase: "100:",
			Balance: 123,
			UserID:  2,
		},
		MachineInfo{
			BetBase: "",
			Balance: 123,
			UserID:  3,
		},
	}

	var newMachineInfo []MachineInfo
	iUserID, err := strconv.Atoi(UserID)
	if err == nil {
		for _, value := range machineInfo {
			if value.UserID == iUserID {
				newMachineInfo = append(newMachineInfo, value)
			}
		}
	}

	//fmt.Println(newMachineInfo)

	// GameDB5.Select("BetBase, Balance, UserID").Where("UserID = ?", UserID).Table("MachineInfo").Find(&machineInfo)

	return newMachineInfo
}

// MachineInfo 機率遊戲的欄位
type MachineInfo struct {
	BetBase string  `gorm:"column:BetBase"`
	Balance float64 `gorm:"column:Balance"`
	UserID  int     `gorm:"column:UserID"`
}
