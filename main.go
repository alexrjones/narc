package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>

const char* GetFrontmostApp();
*/
import "C"
import (
	"fmt"
	"time"
	"unsafe"
)

var lastActive time.Time
var lastApp string

func getFrontApp() string {
	cstr := C.GetFrontmostApp()
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

func main() {
	lastApp = getFrontApp()
	lastActive = time.Now()
	fmt.Println("Starting with app", lastApp)
	for {
		select {
		case <-time.After(10 * time.Second):
			{
				newApp := getFrontApp()
				fmt.Println(newApp, lastApp)
				if newApp != lastApp {
					oldApp := lastApp
					lastApp = newApp
					lastActive = time.Now()
					fmt.Printf("Changed from app %s to app %s\n", oldApp, newApp)
				}
			}
		}
	}
}
