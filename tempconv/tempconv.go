package tempconv

// Celsius 攝氏溫度
type Celsius float64

// Fahrenheit 華氏溫度
type Fahrenheit float64

//CToF 攝氏轉華氏
func CToF(c Celsius) Fahrenheit {
	return Fahrenheit(c*9/5 + 32)
}

//FToC 華氏轉攝氏
func FToC(f Fahrenheit) Celsius {
	return Celsius((f - 32) * 5 / 9)
}
