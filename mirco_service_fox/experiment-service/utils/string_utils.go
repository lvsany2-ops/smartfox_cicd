package utils

import "strconv"

func StrToUint(s string) uint {
	num, _ := strconv.ParseUint(s, 10, 32)
	return uint(num)
}
