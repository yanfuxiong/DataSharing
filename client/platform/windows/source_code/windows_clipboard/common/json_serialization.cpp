#include "json_serialization.h"

void g_globalRegisterForJsonSerialize()
{
    qRegisterMetaType<UpdateDownloadPathMsgPtr>();
    qRegisterMetaType<UpdateClientVersionMsgPtr>();
    qRegisterMetaType<ShowWindowsClipboardMsgPtr>();
    qRegisterMetaType<NotifyErrorEventMsgPtr>();
}
