package tempconv

import (
	"fmt"
)

func main() {

}

// NewTemperature InterFace的產生
func NewTemperature(tp Temperature) Temperature {
	return tp
}

// Temperature 介面可以串接自定義變數接口
type Temperature interface {
	PrintTempFormat() string
}

// Celsius 攝氏溫度
type Celsius float64

// Fahrenheit 華氏溫度
type Fahrenheit float64

// PrintTempFormat 接口
func (c Celsius) PrintTempFormat() string {
	return fmt.Sprintf("%f度C", c)
}

// PrintTempFormat 接口
func (f Fahrenheit) PrintTempFormat() string {
	return fmt.Sprintf("%f度F", f)
}

//CToF 攝氏轉華氏
func CToF(c Celsius) Fahrenheit {
	return Fahrenheit(c*9/5 + 32)
}

//FToC 華氏轉攝氏
func FToC(f Fahrenheit) Celsius {
	return Celsius((f - 32) * 5 / 9)
}
