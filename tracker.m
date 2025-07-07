#import <Cocoa/Cocoa.h>
#import <stdlib.h>

const char* GetFrontmostApp() {
    NSRunningApplication* frontApp = [[NSWorkspace sharedWorkspace] frontmostApplication];
    NSString* bundleID = frontApp.bundleIdentifier ?: @"(unknown)";
    return strdup([bundleID UTF8String]);
}

const void Thing() {
    [[NSWorkspace sharedWorkspace]]
}