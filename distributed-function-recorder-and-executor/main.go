package main

import (
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"
)

type CallBack struct {
	mu        sync.Mutex    // Add a mutex to protect the slice
	Functions []*Function
}

type Function struct {
	Name     string
	Function func()
}

func NewCallBack() *CallBack {
	return &CallBack{
		Functions: make([]*Function, 0),
	}
}

func NewFunction(f func(), name string) *Function {
	return &Function{
		Name:     name,
		Function: f,
	}
}

func (cb *CallBack) record(f func()) {

	// Lock before modifying the slice, and defer the unlock to the end.
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.Functions = append(cb.Functions, NewFunction(f, GetFunctionName(f)))

}

func (cb *CallBack) execute() {

	// We only hold the lock while making this copy.
	cb.mu.Lock()
	funcsCopy := make([]*Function, len(cb.Functions))
	copy(funcsCopy, cb.Functions)
	cb.mu.Unlock() 

	// Execute the functions from the local copy safely, without holding the lock
	for _, functionObject := range funcsCopy {
		fmt.Println("Executing - ", functionObject.Name)
		functionObject.Function()
	}

}

func main() {
	cb := NewCallBack()

	var wg sync.WaitGroup

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int){
			defer wg.Done()
			cb.record(func() { 
				fmt.Println("Hi from func-", i) 
				cb.record(func () {
					fmt.Println("Hi from func-",i,"-child")
				})
			})
			cb.execute()
			time.Sleep(3 * time.Second)
			cb.execute()
		}(i)
	}

	wg.Wait()

}

func GetFunctionName(i interface{}) string {
	// Get the program counter
	pc := reflect.ValueOf(i).Pointer()

	// Get the function object
	f := runtime.FuncForPC(pc)
	if f == nil {
		return ""
	}

	// Return the full name (e.g., "main.foo")
	return f.Name()
}