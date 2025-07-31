package mobileTests

import (
	"reflect"
	"strings"
	"sync"
)

type MobileTests struct{}

func ExecuteMobileTests() {
	tester := &MobileTests{}
	t := reflect.TypeOf(tester)
	v := reflect.ValueOf(tester)
	var wg sync.WaitGroup
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if strings.HasPrefix(method.Name, "TC_") {
			wg.Add(1)
			go func(m reflect.Method) {
				method.Func.Call([]reflect.Value{v})
				wg.Done()
			}(method)
		}
	}
	wg.Wait()
}

func ExecuteSingleMobileTest(testToRun string) {
	tester := &MobileTests{}
	t := reflect.TypeOf(tester)
	v := reflect.ValueOf(tester)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if strings.Contains(method.Name, testToRun) {
			method.Func.Call([]reflect.Value{v})
			return
		}
	}
}

func ExecuteMobileTestsSequentially() {
	tester := &MobileTests{}
	t := reflect.TypeOf(tester)
	v := reflect.ValueOf(tester)
	for i := range t.NumMethod() {
		method := t.Method(i)
		if strings.HasPrefix(method.Name, "TC_") {
			method.Func.Call([]reflect.Value{v})
		}
	}
}
