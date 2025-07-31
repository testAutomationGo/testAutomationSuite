package fullScenarioTests

import (
	"log"
	"reflect"
	"strings"
	"sync"
)

type FullScenarioTests struct{}

func ExecuteFullScenarioTests() {
	FEtester := &FullScenarioTests{}
	t := reflect.TypeOf(FEtester)
	v := reflect.ValueOf(FEtester)
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

func ExecuteSingleFullScenarioTest(testToRun string) {
	FStester := &FullScenarioTests{}
	t := reflect.TypeOf(FStester)
	v := reflect.ValueOf(FStester)
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		if strings.Contains(method.Name, testToRun) {
			method.Func.Call([]reflect.Value{v})
			return
		}
	}
}

func ExecuteFullScenarioTestsSequentially() {
	log.Println("Executing Full Scenario Tests Sequentially")
	tester := &FullScenarioTests{}
	t := reflect.TypeOf(tester)
	v := reflect.ValueOf(tester)
	for i := range t.NumMethod() {
		method := t.Method(i)
		if strings.HasPrefix(method.Name, "TC_") {
			method.Func.Call([]reflect.Value{v})
		}
	}
}
