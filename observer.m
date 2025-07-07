#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>
#include <string.h> // For strdup()
#include <stdlib.h> // For free()

// Forward declaration of the Go function
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
        const char* tempCStr = [bundleIdentifier UTF8String];
        if (tempCStr) {
            cBundleID = strdup(tempCStr);
        }
    }

    goStaticApplicationActivatedCallback((void*)cBundleID);
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

// --- RunLoop Management with a "keep-alive" source ---

// Declare a static CFRunLoopSourceRef to manage its lifecycle.
static CFRunLoopSourceRef dummyRunLoopSource = NULL;

// Dummy callback for the run loop source
void DummyRunLoopSourceCallback(void *info) {
    // This function doesn't need to do anything. Its mere presence keeps the run loop alive.
}

void RunMainRunLoopForever(void) {
    NSLog(@"Objective-C: Main NSRunLoop running forever with keep-alive source.");

    // Get the main run loop
    CFRunLoopRef mainRunLoop = CFRunLoopGetMain();

    // Create a context for the source (optional, but good practice)
    CFRunLoopSourceContext context = {0, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, DummyRunLoopSourceCallback};

    // Create a dummy run loop source that just keeps the loop alive
    // Type 0 source: Custom event source that you signal manually.
    // Order 0: Default order.
    dummyRunLoopSource = CFRunLoopSourceCreate(kCFAllocatorDefault, 0, &context);

    // Add the source to the main run loop in the default mode
    CFRunLoopAddSource(mainRunLoop, dummyRunLoopSource, kCFRunLoopDefaultMode);

    // Run the loop. It will now stay alive because of the source.
    CFRunLoopRun(); // This blocks until CFRunLoopStop is called

    // After CFRunLoopRun returns (i.e., it was stopped), clean up the source
    CFRunLoopRemoveSource(mainRunLoop, dummyRunLoopSource, kCFRunLoopDefaultMode);
    CFRelease(dummyRunLoopSource);
    dummyRunLoopSource = NULL; // Clear the reference

    NSLog(@"Objective-C: Main NSRunLoop finished running.");
}

void StopMainRunLoop(void) {
    NSLog(@"Objective-C: Main NSRunLoop stop requested.");
    CFRunLoopStop(CFRunLoopGetMain());
}
