#ifndef __STRING_UTILS__
#define __STRING_UTILS__

#include <iostream>
#include <vector>
#include <string>
#include <sstream>
#include <map>
#include "typeDefine.h"

std::vector<std::string> SplitString(const std::string& text);
std::string GetValueString(std::stringstream& ss, const std::string& text, std::streampos curPos);
VcpCodeStringVectorMap GetVcpCodeStrMap(const std::string& capabilityString);

#endif // __STRING_UTILS__