package main

//ErrorCode Enum
type ErrorCode int

// iota Init AutoIncrease
const (
	OK           ErrorCode = iota // value --> 0
	DataEmpty                     // value --> 1
	BetBaseEmpty                  // value --> 2
	BetBaseWrong                  // BetBase資料有問題
)

//String fmt
func (errorCode ErrorCode) String() string {

	names := [...]string{
		"OK",
		"DataEmpty",
		"BetBaseEmpty",
		"BetBaseWrong",
	}

	// if errorCode < OK || errorCode > BetBaseEmpty {
	// 	return "Unknown"
	// }

	return names[errorCode]

}
