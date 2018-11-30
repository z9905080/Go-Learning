package main

import (
	"testing"
)

func TestGetMachineGame(t *testing.T) {
	//case 1 2 3 4
	var input string
	input = "1" //OK
	//input = "2" //BetBaseWrong
	//input = "3" //BetBaseEmpty
	//input = "4" //DataEmpty

	returnCode, err := GetMachineGame(input)

	switch returnCode {
	case DataEmpty:
		t.Error("長度為空值, Code:" + DataEmpty.String())
	case BetBaseEmpty:
		t.Error("Bet字串為空值, Code:" + BetBaseEmpty.String())
	case BetBaseWrong:
		t.Error("Bet字串有問題, Code:" + BetBaseWrong.String())
		t.Error(err.Error())
	case OK:
		t.Log("GetMachineGame OK!")
	}

}
