#ifndef CALLBACKS_H
#define CALLBACKS_H

#include <stdlib.h>

typedef void (*GitProgressFunc)(const char* msg, int len, void* userData);
typedef char* (*GitSignFunc)(const char* data, int dataLen, char** sigOut, int* sigLenOut, void* userData);

static inline void callProgressFunc(GitProgressFunc fn, const char* msg, int len, void* userData) {
    fn(msg, len, userData);
}

static inline char* callSignFunc(GitSignFunc fn, const char* data, int dataLen, char** sigOut, int* sigLenOut, void* userData) {
    return fn(data, dataLen, sigOut, sigLenOut, userData);
}

typedef char* (*GitHostKeyFunc)(const char* hostname, const char* remoteAddr,
    const char* keyType, const char* keyBase64, void* userData);

static inline char* callHostKeyFunc(GitHostKeyFunc fn, const char* hostname,
    const char* remoteAddr, const char* keyType, const char* keyBase64, void* userData) {
    return fn(hostname, remoteAddr, keyType, keyBase64, userData);
}

#endif
