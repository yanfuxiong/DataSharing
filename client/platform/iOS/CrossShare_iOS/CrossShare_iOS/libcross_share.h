/* Code generated by cmd/cgo; DO NOT EDIT. */

/* package command-line-arguments */


#line 1 "cgo-builtin-export-prolog"

#include <stddef.h>

#ifndef GO_CGO_EXPORT_PROLOGUE_H
#define GO_CGO_EXPORT_PROLOGUE_H

#ifndef GO_CGO_GOSTRING_TYPEDEF
typedef struct { const char *p; ptrdiff_t n; } _GoString_;
#endif

#endif

/* Start of preamble from import "C" comments.  */


#line 5 "main.go"

#include <stdlib.h>
#include <stdint.h>

typedef void (*CallbackMethodText)(char*);
typedef void (*CallbackMethodImage)(char* content);
typedef void (*LogMessageCallback)(char* msg);
typedef void (*EventCallback)(int event);
typedef void (*CallbackMethodFileConfirm)(char* id, char* platform, char* fileName, long long fileSize);
typedef void (*CallbackMethodFileDone)(char* name, long long fileSize);
typedef void (*CallbackMethodFoundPeer)();
typedef void (*CallbackMethodFileNotify)(char* ip, char* id, char* platform, char* fileName, unsigned long long fileSize,unsigned long long timestamp);
typedef void (*CallbackMethodFileListNotify)(char* ip, char* id, char* platform,unsigned int fileCnt, unsigned long long totalSize,unsigned long long timestamp, char* firstFileName, unsigned long long firstFileSize);
typedef void (*CallbackUpdateProgressBar)(char* id, char* fileName,unsigned long long recvSize,unsigned long long total,unsigned long long timestamp);
typedef void (*CallbackUpdateMultipleProgressBar)(char* ip,char* id, char* deviceName, char* currentfileName,unsigned int recvFileCnt, unsigned int totalFileCnt,unsigned long long currentFileSize,unsigned long long totalSize,unsigned long long recvSize,unsigned long long timestamp);
typedef void (*CallbackFileError)(char* id, char* fileName, char* err);
typedef void (*CallbackMethodStartBrowseMdns)(char* instance, char* serviceType);
typedef void (*CallbackMethodStopBrowseMdns)();
typedef char* (*CallbackAuthData)();

static CallbackMethodText gCallbackMethodText = 0;
static CallbackMethodImage gCallbackMethodImage = 0;
static LogMessageCallback gLogMessageCallback = 0;
static EventCallback gEventCallback = 0;
static CallbackMethodFileConfirm gCallbackMethodFileConfirm = 0;
static CallbackMethodFileDone gCallbackMethodFileDone = 0;
static CallbackMethodFoundPeer gCallbackMethodFoundPeer = 0;
static CallbackMethodFileNotify gCallbackMethodFileNotify = 0;
static CallbackMethodFileListNotify gCallbackMethodFileListNotify = 0;
static CallbackUpdateProgressBar gCallbackUpdateProgressBar = 0;
static CallbackUpdateMultipleProgressBar gCallbackUpdateMultipleProgressBar = 0;
static CallbackFileError gCallbackFileError = 0;
static CallbackMethodStartBrowseMdns gCallbackMethodStartBrowseMdns = 0;
static CallbackMethodStopBrowseMdns gCallbackMethodStopBrowseMdns = 0;
static CallbackAuthData gCallbackAuthData = 0;

static void setCallbackMethodText(CallbackMethodText cb) {gCallbackMethodText = cb;}
static void invokeCallbackMethodText(char* str) {
	if (gCallbackMethodText) {gCallbackMethodText(str);}
}
static void setCallbackMethodImage(CallbackMethodImage cb) {gCallbackMethodImage = cb;}
static void invokeCallbackMethodImage(char* str) {
	if (gCallbackMethodImage) {gCallbackMethodImage(str);}
}
static void setLogMessageCallback(LogMessageCallback cb) {gLogMessageCallback = cb;}
static void invokeLogMessageCallback(char* str) {
	if (gLogMessageCallback) {gLogMessageCallback(str);}
}
static void setEventCallback(EventCallback cb) {gEventCallback = cb;}
static void invokeEventCallback(int event) {
	if (gEventCallback) {gEventCallback(event);}
}
static void setCallbackMethodFileConfirm(CallbackMethodFileConfirm cb) {gCallbackMethodFileConfirm = cb;}
static void invokeCallbackMethodFileConfirm(char* id, char* platform, char* fileName, long long fileSize) {
	if (gCallbackMethodFileConfirm) {gCallbackMethodFileConfirm(id, platform, fileName, fileSize);}
}
static void setCallbackMethodFileDone(CallbackMethodFileDone cb) {gCallbackMethodFileDone = cb;}
static void invokeCallbackMethodFileDone(char* name, long long fileSize) {
	if (gCallbackMethodFileDone) {gCallbackMethodFileDone(name, fileSize);}
}
static void setCallbackMethodFoundPeer(CallbackMethodFoundPeer cb) {gCallbackMethodFoundPeer = cb;}
static void invokeCallbackMethodFoundPeer() {
	if (gCallbackMethodFoundPeer) {gCallbackMethodFoundPeer();}
}
static void setCallbackMethodFileNotify(CallbackMethodFileNotify cb) {gCallbackMethodFileNotify = cb;}
static void invokeCallbackMethodFileNotify(char* ip, char* id, char* platform, char* fileName, unsigned long long fileSize,unsigned long long timestamp) {
	if (gCallbackMethodFileNotify) {gCallbackMethodFileNotify(ip, id, platform, fileName, fileSize, timestamp);}
}
static void setCallbackMethodFileListNotify(CallbackMethodFileListNotify cb) {gCallbackMethodFileListNotify = cb;}
static void invokeCallbackMethodFileListNotify(char* ip, char* id, char* platform,unsigned int fileCnt, unsigned long long totalSize,unsigned long long timestamp, char* firstFileName, unsigned long long firstFileSize) {
	if (gCallbackMethodFileListNotify) {gCallbackMethodFileListNotify(ip, id, platform, fileCnt, totalSize, timestamp, firstFileName, firstFileSize);}
}
static void setCallbackUpdateProgressBar(CallbackUpdateProgressBar cb) {gCallbackUpdateProgressBar = cb;}
static void invokeCallbackUpdateProgressBar(char* id, char* fileName,unsigned long long recvSize,unsigned long long total,unsigned long long timestamp) {
	if (gCallbackUpdateProgressBar) {gCallbackUpdateProgressBar(id, fileName, recvSize, total,timestamp);}
}
static void setCallbackUpdateMultipleProgressBar(CallbackUpdateMultipleProgressBar cb) {gCallbackUpdateMultipleProgressBar = cb;}
static void invokeCallbackUpdateMultipleProgressBar(char* ip,char* id, char* deviceName, char* currentfileName,unsigned int recvFileCnt, unsigned int totalFileCnt,unsigned long long currentFileSize,unsigned long long totalSize,unsigned long long recvSize,unsigned long long timestamp) {
	if (gCallbackUpdateMultipleProgressBar) {gCallbackUpdateMultipleProgressBar(ip,id, deviceName,currentfileName,recvFileCnt,totalFileCnt,currentFileSize,totalSize, recvSize, timestamp);}
}
static void setCallbackFileError(CallbackFileError cb) {gCallbackFileError = cb;}
static void invokeCallbackFileError(char* id, char* fileName, char* err) {
	if (gCallbackFileError) {gCallbackFileError(id, fileName, err);}
}
static void setCallbackMethodStartBrowseMdns(CallbackMethodStartBrowseMdns cb) {gCallbackMethodStartBrowseMdns = cb;}
static void invokeCallbackMethodStartBrowseMdns(char* instance, char* serviceType) {
	if (gCallbackMethodStartBrowseMdns) {gCallbackMethodStartBrowseMdns(instance, serviceType);}
}
static void setCallbackMethodStopBrowseMdns(CallbackMethodStopBrowseMdns cb) {gCallbackMethodStopBrowseMdns = cb;}
static void invokeCallbackMethodStopBrowseMdns() {
	if (gCallbackMethodStopBrowseMdns) {gCallbackMethodStopBrowseMdns();}
}
static void setCallbackGetAuthData(CallbackAuthData cb) {gCallbackAuthData = cb;}
static char* invokeCallbackGetAuthData() {
	if (gCallbackAuthData) { return gCallbackAuthData();}
	return NULL;
}

#line 1 "cgo-generated-wrapper"


/* End of preamble from import "C" comments.  */


/* Start of boilerplate cgo prologue.  */
#line 1 "cgo-gcc-export-header-prolog"

#ifndef GO_CGO_PROLOGUE_H
#define GO_CGO_PROLOGUE_H

typedef signed char GoInt8;
typedef unsigned char GoUint8;
typedef short GoInt16;
typedef unsigned short GoUint16;
typedef int GoInt32;
typedef unsigned int GoUint32;
typedef long long GoInt64;
typedef unsigned long long GoUint64;
typedef GoInt64 GoInt;
typedef GoUint64 GoUint;
typedef size_t GoUintptr;
typedef float GoFloat32;
typedef double GoFloat64;
#ifdef _MSC_VER
#include <complex.h>
typedef _Fcomplex GoComplex64;
typedef _Dcomplex GoComplex128;
#else
typedef float _Complex GoComplex64;
typedef double _Complex GoComplex128;
#endif

/*
  static assertion to make sure the file is being used on architecture
  at least with matching size of GoInt.
*/
typedef char _check_for_64_bit_pointer_matching_GoInt[sizeof(void*)==64/8 ? 1:-1];

#ifndef GO_CGO_GOSTRING_TYPEDEF
typedef _GoString_ GoString;
#endif
typedef void *GoMap;
typedef void *GoChan;
typedef struct { void *t; void *v; } GoInterface;
typedef struct { void *data; GoInt len; GoInt cap; } GoSlice;

#endif

/* End of boilerplate cgo prologue.  */

#ifdef __cplusplus
extern "C" {
#endif

extern void SetCallbackMethodText(CallbackMethodText cb);
extern void SetCallbackMethodImage(CallbackMethodImage cb);
extern void SetLogMessageCallback(LogMessageCallback cb);
extern void SetEventCallback(EventCallback cb);
extern void SetCallbackMethodFileConfirm(CallbackMethodFileConfirm cb);
extern void SetCallbackMethodFileDone(CallbackMethodFileDone cb);
extern void SetCallbackMethodFoundPeer(CallbackMethodFoundPeer cb);
extern void SetCallbackMethodFileNotify(CallbackMethodFileNotify cb);
extern void SetCallbackMethodFileListNotify(CallbackMethodFileListNotify cb);
extern void SetCallbackUpdateProgressBar(CallbackUpdateProgressBar cb);
extern void SetCallbackUpdateMultipleProgressBar(CallbackUpdateMultipleProgressBar cb);
extern void SetCallbackFileError(CallbackFileError cb);
extern void SetCallbackMethodStartBrowseMdns(CallbackMethodStartBrowseMdns cb);
extern void SetCallbackMethodStopBrowseMdns(CallbackMethodStopBrowseMdns cb);
extern void MainInit(GoString deviceName, GoString serverId, GoString serverIpInfo, GoString listenHost, GoInt listenPort);
extern void SendText(GoString s);
extern char* GetClientList();
extern void SendImage(GoString content);
extern void SendAddrsFromPlatform(GoString addrsList);
extern void SendNetInterfaces(GoString name, GoString mac, GoInt mtu, GoInt index, GoUint flag);
extern void SendFileDropRequest(GoString filePath, GoString id, GoInt64 fileSize);
extern void SendMultiFilesDropRequest(GoString multiFilesData);
extern void SetFileDropResponse(GoString fileName, GoString id, GoUint8 isReceive);
extern void SetNetWorkConnected(GoUint8 isConnect);
extern void SetHostListenAddr(GoString listenHost, GoInt listenPort);
extern void SetDIASID(GoString DiasID);
extern void SetDetectPluginEvent(GoUint8 isPlugin);
extern void SetCallbackGetAuthData(CallbackAuthData cb);
extern char* GetVersion();
extern char* GetBuildDate();
extern void SetupRootPath(GoString rootPath);
extern void SetSrcAndPort(GoInt source, GoInt port);
extern void SetBrowseMdnsResult(GoString instance, GoString ip, GoInt port);
extern void SetConfirmDocumentsAccept(GoUint8 ifConfirm);
extern void FreeCString(char* p);

#ifdef __cplusplus
}
#endif
