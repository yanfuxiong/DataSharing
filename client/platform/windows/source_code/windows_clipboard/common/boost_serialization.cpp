#include "boost_serialization.h"

void g_globalRegisterForBoostSerialize()
{
    qRegisterMetaType<SystemConfigPtr>();
    qRegisterMetaType<UpdateLocalConfigInfoMsgPtr>();
    qRegisterMetaType<UpdateClientStatusMsgPtr>();
}
