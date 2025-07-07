#ifndef SLEEPWATCHER_H
#define SLEEPWATCHER_H
void StartSleepWatcher(void (*onSleep)(void), void (*onWake)(void));
#endif
