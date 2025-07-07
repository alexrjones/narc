//go:build darwin

package main

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include <CoreFoundation/CoreFoundation.h>
#include <IOKit/IOKitLib.h>
#include <IOKit/IOCFPlugIn.h>
#include <IOKit/hidsystem/IOHIDLib.h>
#include <IOKit/hidsystem/IOHIDParameter.h>
#include <IOKit/hidsystem/event_status_driver.h>
float GetIdleTimeSeconds() {
    io_iterator_t iterator;
    io_registry_entry_t entry;
    kern_return_t kr;
    CFMutableDictionaryRef dict = IOServiceMatching("IOHIDSystem");
    kr = IOServiceGetMatchingServices(0, dict, &iterator);
    if (kr != KERN_SUCCESS) {
        return -1;
    }
    entry = IOIteratorNext(iterator);
    IOObjectRelease(iterator);
    if (entry == 0) {
        return -1;
    }
    CFTypeRef obj = IORegistryEntryCreateCFProperty(entry,
                CFSTR("HIDIdleTime"), kCFAllocatorDefault, 0);
    IOObjectRelease(entry);
    if (!obj) {
        return -1;
    }
    int64_t nanoseconds = 0;
    if (CFGetTypeID(obj) == CFDataGetTypeID()) {
        CFDataGetBytes((CFDataRef)obj, CFRangeMake(0, sizeof(nanoseconds)),
                       (UInt8*)&nanoseconds);
    } else if (CFGetTypeID(obj) == CFNumberGetTypeID()) {
        CFNumberGetValue((CFNumberRef)obj, kCFNumberSInt64Type, &nanoseconds);
    } else {
        CFRelease(obj);
        return -1;
    }
    CFRelease(obj);
    return nanoseconds / 1.0e9; // convert nanoseconds to seconds
}
*/
import "C"

func getIdleSeconds() float64 {
	return float64(C.GetIdleTimeSeconds())
}
