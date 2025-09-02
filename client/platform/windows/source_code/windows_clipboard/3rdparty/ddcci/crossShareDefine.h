#ifndef __CROSS_SHARE_DEFINE__
#define __CROSS_SHARE_DEFINE__

#include <string>
#include <vector>
#include <map>

#define DDCCI_CrossShare_CMD_AUTH_DEVICE                0xE0 // do set then get
#define DDCCI_CrossShare_CMD_GET_MAC_1                  0xE1 // read only
#define DDCCI_CrossShare_CMD_GET_MAC_2                  0xE2 // read only
#define DDCCI_CrossShare_CMD_GET_TV_SRC                 0xE3 // read only
#define DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_W   0xE4 // write only
#define DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_H   0xE5 // write only
#define DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_X      0xE6 // write only
#define DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_Y      0xE7 // write only
#define DDCCI_CrossShare_CMD_GET_CUSTOMIZED_THEME       0xE8 // read only

#define CUSTOMER_ID 44 // ASUS

enum CrossShareVcpCmdEnum
{
    CS_VCP_AUTH_DEVICE,
    CS_VCP_GET_MAC_ADDR_PART1,
    CS_VCP_GET_MAC_ADDR_PART2,
    CS_VCP_GET_TV_SRC,
    CS_VCP_SET_DESKTOP_RESOLUTION_W,
    CS_VCP_SET_DESKTOP_RESOLUTION_H,
    CS_VCP_SET_CURSOR_POSITION_X,
    CS_VCP_SET_CURSOR_POSITION_Y,
    CS_VCP_GET_CUSTOMIZED_THEME,
};

const std::map<BYTE, std::pair<std::string, std::string>> kCrossShareVcpCodeMap = {
//   vcpCode,                               {"get", "set"}
    {DDCCI_CrossShare_CMD_AUTH_DEVICE,      {"Reply Auth Status", "Query Auth"}},
    {DDCCI_CrossShare_CMD_GET_MAC_1,        {"Get Mac Addr part1", "not support"}},
    {DDCCI_CrossShare_CMD_GET_MAC_2,        {"Get Mac Addr part2", "not support"}},
    {DDCCI_CrossShare_CMD_GET_TV_SRC,       {"Get TvSource", "not support"}},
    {DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_W,  {"not support", "set PC width"}},
    {DDCCI_CrossShare_CMD_SET_DESKTOP_RESOLUTION_H,  {"not support", "set PC hight"}},
    {DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_X,     {"not support", "set PC posX"}},
    {DDCCI_CrossShare_CMD_SET_CURSOR_POSITION_Y,     {"not support", "set PC posY"}},
    {DDCCI_CrossShare_CMD_GET_CUSTOMIZED_THEME,      {"Get ThemeCode", "not support"}}
};

struct CrossShareThemeCode {
    union
    {
      struct
      {
        DWORD CustomerId : 10;
        DWORD StyleId    : 6;
        DWORD Reserved   : 16;
      };
      unsigned char byte[4];
    };   
};
#endif // __CROSS_SHARE_DEFINE__