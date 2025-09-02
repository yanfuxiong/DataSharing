#include "strUtils.h"

using namespace std;

std::vector<std::string> SplitString(const std::string& text) {
    std::vector<std::string> words;
    std::istringstream ss(text);
    std::string token;

    while (ss >> token) {
        words.push_back(token);
    }

    return words;
}

string GetValueString(stringstream& ss, const string& text, streampos curPos)
{
    string contextTail = text.substr(curPos);
    size_t rBracket = contextTail.find(")");
    size_t lBracket = contextTail.find("(");
    ss.seekg(curPos + static_cast<std::streampos>(rBracket+1), std::ios::beg);

    contextTail = contextTail.substr(lBracket+1);
    rBracket = contextTail.find(")");

    return contextTail.erase(rBracket);
}

VcpCodeStringVectorMap ParseVcpCode(const std::string& vcpOpCodesContext)
{
    VcpCodeStringVectorMap vcpMap;

    std::stringstream ss(vcpOpCodesContext);
    std::string lastKey = "";
    std::string token;

    while (ss >> token)
    {
        bool nonContinuesTypeVcp = false;
        std::streampos curPos = ss.tellg();
        size_t dirtyTokenPos = token.find("(");

        std::string followedToken = "";
        ss >> followedToken;
        ss.seekg(curPos, std::ios::beg); //recover curPos

        if (dirtyTokenPos != string::npos) {
            // ex: "14(04 05 08 0B)"
            nonContinuesTypeVcp = true;
            size_t dirtyTokenLen = token.size();
            string key = token.erase(dirtyTokenPos);

            size_t strShift = dirtyTokenLen - dirtyTokenPos;
            ss.seekg(curPos - static_cast<std::streampos>(strShift), std::ios::beg);
            curPos = ss.tellg(); // update pos, seek to "("

            string valueStr = GetValueString(ss, vcpOpCodesContext, curPos);
            vector<std::string> valueStrVec = SplitString(valueStr);

            vcpMap[key] = valueStrVec;
            // std::cout << "case1: " << key << " " << valueStr << endl;
        } else if (followedToken.find("(") == 0) {
            // ex: "14 (04 05 08 0B)"
            nonContinuesTypeVcp = true;
            string key = token;

            string valueStr = GetValueString(ss, vcpOpCodesContext, curPos);
            vector<std::string> valueStrVec = SplitString(valueStr);

            vcpMap[key] = valueStrVec;
            // std::cout << "case2: " << key << " " << valueStr << endl;
        } else {
            vcpMap[token] = {};
        }
    }

    return vcpMap;
}

std::string ExtractVCPOpCodesString(const std::string& capabilityString)
{
    std::string vcpString;
    size_t pos = capabilityString.find("vcp");
    if (pos == std::string::npos) {
        return "";
    }
    vcpString = capabilityString.substr(pos);

    size_t start = vcpString.find('(');  // Find the first '('
    if (start == std::string::npos) return "";

    int bracketCount = 1;  // Track the bracket nesting level
    size_t end = start + 1;

    // Break the while loop when the matching r-bracket for the first l-bracket is found.
    while (end < vcpString.size() && bracketCount > 0) {
        if (vcpString[end] == '(') {
            bracketCount++;
        } else if (vcpString[end] == ')') {
            bracketCount--;
        }
        end++;
    }

    // return the content of "vcp(...)"
    if (bracketCount == 0) {
        return vcpString.substr(start + 1, (end - start - 2));  // remove the first l-bracket and the lastest r-bracket.
    }

    return "";  // Parsing failed
}

VcpCodeStringVectorMap GetVcpCodeStrMap(const std::string& capabilityString)
{
    std::string vcpContent = ExtractVCPOpCodesString(capabilityString);
    return ParseVcpCode(vcpContent);
}