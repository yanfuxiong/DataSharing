#ifndef __INCLUDED_MS_COMMON_EXT__
#define __INCLUDED_MS_COMMON_EXT__

#ifdef __cplusplus
extern "C" {
#endif

typedef enum EVENT_TYPE
{
    EVENT_TYPE_OPEN_FILE_ERR = 0,
    EVENT_TYPE_RECV_TIMEOUT,
} EVENT_TYPE;

typedef enum FILE_DROP_CMD
{
    FILE_DROP_REQUEST = 0,
    FILE_DROP_ACCEPT,
    FILE_DROP_REJECT,
} FILE_DROP_CMD;

typedef enum NOTI_MSG_CODE
{
    NOTI_MSG_CODE_CONN_STATUS_SUCCESS       = 1,
    NOTI_MSG_CODE_CONN_STATUS_FAIL          = 2,
    NOTI_MSG_CODE_FILE_TRANS_DONE_SENDER    = 3,
    NOTI_MSG_CODE_FILE_TRANS_DONE_RECEIVER  = 4,
    NOTI_MSG_CODE_FILE_TRANS_REJECT         = 5,
} NOTI_MSG_CODE;

typedef struct IMAGE_HEADER
{
    int width;
    int height;
    unsigned short planes;
    unsigned short bitCount;
    unsigned long compression;
} IMAGE_HEADER;

#ifdef __cplusplus
}
#endif

#endif //__INCLUDED_MS_COMMON_EXT__