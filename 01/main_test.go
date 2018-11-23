package main

import (
	"testing"
)

func TestGetMachineGame(t *testing.T) {
	returnCode := GetMachineGame("0")

	switch returnCode {
	case DataEmpty:
		t.Error("長度為空值, Code:" + DataEmpty.String())
	case BetBaseEmpty:
		t.Error("Bet字串為空值, Code:" + BetBaseEmpty.String())
	case BetBaseWrong:
		t.Error("Bet字串有問題, Code:" + BetBaseWrong.String())
	case OK:
		t.Log("GetMachineGame OK!")
	}

}
