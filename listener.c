#include "listener.h"
#include <ApplicationServices/ApplicationServices.h>
#include <CoreFoundation/CoreFoundation.h>

static void (*activityCallback)(void) = 0;
static CFMachPortRef eventTap = NULL;
static CFRunLoopSourceRef runLoopSource = NULL;

CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type,
                            CGEventRef event, void *refcon) {
    if (type == kCGEventKeyDown || type == kCGEventMouseMoved ||
        type == kCGEventLeftMouseDown || type == kCGEventRightMouseDown) {
        if (activityCallback) {
            activityCallback();
        }
    }
    return event;
}

void StartEventTap(void (*callback)(void)) {
    activityCallback = callback;
    CGEventMask mask = (1 << kCGEventKeyDown) |
                       (1 << kCGEventMouseMoved) |
                       (1 << kCGEventLeftMouseDown) |
                       (1 << kCGEventRightMouseDown);
    eventTap = CGEventTapCreate(kCGSessionEventTap,
                                              kCGHeadInsertEventTap,
                                              kCGEventTapOptionDefault,
                                              mask,
                                              eventTapCallback,
                                              NULL);
    if (!eventTap) {
        fprintf(stderr, "Failed to create event tap\n");
        return;
    }
    runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTap, true);
    CFRunLoopRun();
}

void StopEventTap() {
    if (eventTap) {
        CGEventTapEnable(eventTap, false);
        if (runLoopSource) {
            CFRunLoopRemoveSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
            CFRelease(runLoopSource);
            runLoopSource = NULL;
        }
        CFMachPortInvalidate(eventTap);
        CFRelease(eventTap);
        eventTap = NULL;
    }
    CFRunLoopStop(CFRunLoopGetCurrent());
}
