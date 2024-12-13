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