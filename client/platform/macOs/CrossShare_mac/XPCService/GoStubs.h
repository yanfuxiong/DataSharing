//
//  GoStubs.h
//  CrossShareHelper
//
//  Created by TS on 2025/8/27.
//  Go 函数声明头文件
//

#ifndef GoStubs_h
#define GoStubs_h

#include <stdio.h>

typedef struct GoResult {
    int success;
    char* data;
    char* error_message;
} GoResult;

GoResult* StartCrossShareService(const char* config);
GoResult* StopCrossShareService(void);
GoResult* UpdateCrossShareConfig(const char* config);

GoResult* StartDeviceDiscovery(void);
GoResult* StopDeviceDiscovery(void);
GoResult* GetDiscoveredDevices(void);

GoResult* ConnectToDevice(const char* device_id);
GoResult* DisconnectFromDevice(const char* device_id);

GoResult* SendFile(const char* file_path, const char* device_id);
GoResult* SendFolder(const char* folder_path, const char* device_id);
GoResult* CancelTransfer(const char* transfer_id);

GoResult* GetLocalIPAddress(void);
GoResult* CheckPortAvailability(int port);

GoResult* HealthCheck(void);

#endif /* GoStubs_h */
