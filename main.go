package main

/*
#cgo CFLAGS: -x objective-c -DNS_BUILD_32_BIT_ENUMS_UNAVAILABLE=1
#cgo LDFLAGS: -framework Foundation -framework AppKit

#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>

// Objective-C functions to register/unregister the static observer
extern void RegisterObjectiveCObserver(void);
extern void UnregisterObjectiveCObserver(void);

// Objective-C functions to manage the NSRunLoop
extern void RunMainRunLoopForever(void);
extern void StopMainRunLoop(void);

// Standard C library for free()
#include <stdlib.h> // For free()

*/
import "C"
import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"unsafe"
)

//export goStaticApplicationActivatedCallback
func goStaticApplicationActivatedCallback(bundleIDStrPtr unsafe.Pointer) {
	// Lock the OS thread for Objective-C runtime interactions (good practice)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// IMPORTANT: The C string received here was allocated with strdup() on the Objective-C side.
	// We are now responsible for freeing it.
	defer C.free(bundleIDStrPtr) // Ensure memory is freed when the Go function exits

	var bundleID string
	if bundleIDStrPtr != nil {
		bundleID = C.GoString((*C.char)(bundleIDStrPtr))
	}

	fmt.Printf("Go Callback: Application Activated (Bundle ID: %s)\n", bundleID)
}

// Public Go API to manage the observer
func RegisterGoObserverForAppActivation() {
	C.RegisterObjectiveCObserver()
}

func UnregisterGoObserverForAppActivation() {
	C.UnregisterObjectiveCObserver()
}

func main() {
	fmt.Println("Go: Starting app...")

	// 1. Register the Objective-C observer (which will attach to the main thread's run loop)
	RegisterGoObserverForAppActivation()
	fmt.Println("Go: Registered for NSWorkspaceDidActivateApplicationNotification.")

	// 2. Set up signal handling in a separate goroutine
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM) // Listen for Ctrl+C and termination signals

	// This goroutine will listen for signals and then tell the main run loop to stop.
	go func() {
		<-sigChan // Blocks until a signal is received
		fmt.Println("\nGo: Termination signal received. Stopping main NSRunLoop...")
		C.StopMainRunLoop() // Signal the main run loop to stop
	}()

	// 3. Run the main thread's NSRunLoop (this call blocks the main goroutine)
	// This goroutine is already on the main OS thread and is implicitly locked.
	fmt.Println("Go: Running main NSRunLoop. Press Ctrl+C to exit.")
	C.RunMainRunLoopForever() // This call blocks until C.StopMainRunLoop() is called

	// 4. Once RunMainRunLoopForever returns (due to StopMainRunLoop being called)
	fmt.Println("Go: Main NSRunLoop stopped.")

	// 5. Unregister the observer
	UnregisterGoObserverForAppActivation()
	fmt.Println("Go: Unregistered from notifications.")

	fmt.Println("Go: Exiting app.")
}
