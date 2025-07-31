package testingToolkit

import (
	"math/rand"
	"time"
)

func GetAlNumString(n int) string {
	alphaNumericString := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"
	alNum := ""
	for i := 0; i < n; i++ {
		index := rand.Intn(len(alphaNumericString))
		alNum += string(alphaNumericString[index])
	}
	return alNum
}

func GetNumericString(sizeOfString int) string {
	numericString := "0123456789"
	numString := ""
	for i := 0; i < sizeOfString; i++ {
		index := rand.Intn(len(numericString))
		numString += string(numericString[index])
	}
	return numString
}

func GetNumericStringWithoutZero(sizeOfString int) string {
	numericString := "123456789"
	numString := ""
	for i := 0; i < sizeOfString; i++ {
		index := rand.Intn(len(numericString))
		numString += string(numericString[index])
	}
	return numString
}

func GetLowerAlphaString(n int) string {
	alphaString := "abcdefghijklmnopqrstuvwxyz"
	alphaStringResult := ""
	for i := 0; i < n; i++ {
		index := rand.Intn(len(alphaString))
		alphaStringResult += string(alphaString[index])
	}
	return alphaStringResult
}

func GetLowerAlphaNumString(n int) string {
	alphaNumericString := "abcdefghijklmopqrtuvwyz012456789"
	alphaNumStringResult := ""
	for i := 0; i < n; i++ {
		index := rand.Intn(len(alphaNumericString))
		alphaNumStringResult += string(alphaNumericString[index])
	}

	return alphaNumStringResult
}

func GenerateRand255Number() string {
	rand.New(rand.NewSource(time.Now().UnixNano()))
	return ConvertIntToString(rand.Intn(255) + 1)
}
