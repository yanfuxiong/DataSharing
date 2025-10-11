//
//  GoStubs.c
//  CrossShareHelper
//
//  Created by TS on 2025/8/27.
//

#include <stdlib.h>
#include <string.h>

typedef struct {
    int success;
    char* data;
    char* error_message;
} GoResult;

GoResult* create_success_result(const char* data) {
    GoResult* result = (GoResult*)malloc(sizeof(GoResult));
    result->success = 1;
    result->data = data ? strdup(data) : NULL;
    result->error_message = NULL;
    return result;
}

GoResult* create_error_result(const char* error) {
    GoResult* result = (GoResult*)malloc(sizeof(GoResult));
    result->success = 0;
    result->data = NULL;
    result->error_message = error ? strdup(error) : strdup("Unknown error");
    return result;
}

GoResult* StartCrossShareService(const char* config) {
    return create_success_result("Service started (stub)");
}

GoResult* StopCrossShareService(void) {
    return create_success_result("Service stopped (stub)");
}

GoResult* UpdateCrossShareConfig(const char* config) {
    return create_success_result("Config updated (stub)");
}

GoResult* StartDeviceDiscovery(void) {
    return create_success_result("Device discovery started (stub)");
}

GoResult* StopDeviceDiscovery(void) {
    return create_success_result("Device discovery stopped (stub)");
}

GoResult* GetDiscoveredDevices(void) {
    const char* mock_devices = "[{\"id\":\"mock-device-1\",\"name\":\"Mock Device\",\"ipAddress\":\"192.168.1.100\",\"port\":8080,\"deviceType\":\"computer\",\"platform\":\"macOS\",\"isOnline\":true,\"lastSeen\":1693123456,\"capabilities\":[\"file_transfer\",\"clipboard_sync\"]}]";
    return create_success_result(mock_devices);
}

GoResult* ConnectToDevice(const char* device_id) {
    return create_success_result("Connected (stub)");
}

GoResult* DisconnectFromDevice(const char* device_id) {
    return create_success_result("Disconnected (stub)");
}

GoResult* SendFile(const char* file_path, const char* device_id) {
    return create_success_result("transfer-id-12345");
}

GoResult* SendFolder(const char* folder_path, const char* device_id) {
    return create_success_result("transfer-id-67890");
}

GoResult* CancelTransfer(const char* transfer_id) {
    return create_success_result("Transfer cancelled (stub)");
}

GoResult* GetLocalIPAddress(void) {
    return create_success_result("192.168.1.50");
}

GoResult* CheckPortAvailability(int port) {
    return create_success_result("1");
}

GoResult* HealthCheck(void) {
    return create_success_result("{\"status\":\"healthy\",\"uptime\":3600}");
}
