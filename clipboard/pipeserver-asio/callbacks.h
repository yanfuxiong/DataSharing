#pragma once
#include "timestamp.h"
#include "global.h"

namespace sunkang {

class PipeConnection;
class Buffer;
typedef std::shared_ptr<PipeConnection> PipeConnectionPtr;
typedef std::function<void (const PipeConnectionPtr&)> ConnectionCallback;
typedef std::function<void (const PipeConnectionPtr&)> CloseCallback;
typedef std::function<void (const PipeConnectionPtr&)> WriteCompleteCallback;

typedef std::function<void (const PipeConnectionPtr&, Buffer*)> MessageCallback;

void defaultConnectionCallback(const PipeConnectionPtr &conn);
void defaultMessageCallback(const PipeConnectionPtr &conn, Buffer*);

}
