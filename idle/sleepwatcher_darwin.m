#import <IOKit/IOMessage.h>
#import <IOKit/pwr_mgt/IOPMLib.h>
#import <IOKit/IOKitLib.h>
#include <stdio.h>
#include "sleepwatcher_darwin.h"

static void (*sleepCallback)(void) = NULL;
static void (*wakeCallback)(void) = NULL;
static io_connect_t root_port;
static IONotificationPortRef notifyPortRef;
static io_object_t notifier;
void sleepCallbackProxy(void *refCon, io_service_t service, natural_t messageType, void *messageArgument) {
    switch (messageType) {
        case kIOMessageCanSystemSleep:
            // Allow sleep
            IOAllowPowerChange(root_port, (long)messageArgument);
            break;
        case kIOMessageSystemWillSleep:
            if (sleepCallback) sleepCallback();
            IOAllowPowerChange(root_port, (long)messageArgument);
            break;
        case kIOMessageSystemHasPoweredOn:
            if (wakeCallback) wakeCallback();
            break;
    }
}
void StartSleepWatcher(void (*onSleep)(void), void (*onWake)(void)) {
    sleepCallback = onSleep;
    wakeCallback = onWake;
    root_port = IORegisterForSystemPower(0, &notifyPortRef, sleepCallbackProxy, &notifier);
    if (root_port == 0) {
        fprintf(stderr, "IORegisterForSystemPower failed\n");
        return;
    }
    CFRunLoopAddSource(CFRunLoopGetCurrent(),
        IONotificationPortGetRunLoopSource(notifyPortRef),
        kCFRunLoopCommonModes);
    CFRunLoopRun();
}