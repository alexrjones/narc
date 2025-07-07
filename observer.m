#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#include <string.h> // For strdup()
#include <stdlib.h> // For free() (though not used for returned strdup, good to include)

// Forward declaration of the Go function
// It now expects a char* (from strdup)
extern void goStaticApplicationActivatedCallback(void* bundleIDStrPtr);

// Define the static Objective-C observer instance
@interface StaticAppObserver : NSObject
- (void)handleApplicationActivated:(NSNotification*)notification;
@end

@implementation StaticAppObserver

- (void)handleApplicationActivated:(NSNotification*)notification {
    NSRunningApplication *activatedApp = notification.userInfo[NSWorkspaceApplicationKey];

    NSString* bundleIdentifier = [activatedApp bundleIdentifier];

    char* cBundleID = NULL;
    if (bundleIdentifier) {
        // Get the UTF8String C string from NSString
        const char* tempCStr = [bundleIdentifier UTF8String];
        if (tempCStr) {
            // Duplicate the string. Go will be responsible for freeing this memory.
            cBundleID = strdup(tempCStr);
        }
    }

    goStaticApplicationActivatedCallback((void*)cBundleID);
    // DO NOT free cBundleID here; Go is responsible for it now.
}

- (void)dealloc {
    NSLog(@"Objective-C: StaticAppObserver deallocating.");
    [super dealloc];
}

@end

// Global static instance of our observer.
static StaticAppObserver* gStaticAppObserver = nil;
static dispatch_once_t observerOnceToken;

// --- Exposed functions for Go to call (Observer management) ---

void RegisterObjectiveCObserver(void) {
    dispatch_once(&observerOnceToken, ^{
        gStaticAppObserver = [[StaticAppObserver alloc] init];
    });

    [[[NSWorkspace sharedWorkspace] notificationCenter] addObserver:gStaticAppObserver
                                                           selector:@selector(handleApplicationActivated:)
                                                               name:NSWorkspaceDidActivateApplicationNotification
                                                             object:nil];
    NSLog(@"Objective-C: Registered static NSNotificationCenter observer.");
}

void UnregisterObjectiveCObserver(void) {
    if (gStaticAppObserver) {
        [[[NSWorkspace sharedWorkspace] notificationCenter] removeObserver:gStaticAppObserver
                                                                      name:NSWorkspaceDidActivateApplicationNotification
                                                                    object:nil];
        NSLog(@"Objective-C: Unregistered static NSNotificationCenter observer.");
    } else {
        NSLog(@"Objective-C: Warning: Tried to unregister non-existent static observer.");
    }
}

// --- Exposed functions for Go to call (NSRunLoop management) ---

void RunRunLoopForever(void) {
    NSLog(@"Objective-C: NSRunLoop running forever.");
    [[NSRunLoop currentRunLoop] run];
    NSLog(@"Objective-C: NSRunLoop finished running.");
}

void StopRunLoop(void) {
    NSLog(@"Objective-C: NSRunLoop stop requested.");
    CFRunLoopStop(CFRunLoopGetCurrent());
}
