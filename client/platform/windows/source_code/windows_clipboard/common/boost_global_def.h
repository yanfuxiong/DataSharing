#pragma once
#include <cstdio>
#include <ctime>

// add_definitions(-DBOOST_DATE_TIME_HAS_REENTRANT_STD_FUNCTIONS)
std::tm *localtime_r(const std::time_t *t, std::tm *result);
std::tm *gmtime_r(const std::time_t *t, std::tm *result);
