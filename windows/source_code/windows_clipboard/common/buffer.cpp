#include "buffer.h"

const size_t Buffer::kCheapPrepend = 8;
const size_t Buffer::kInitialSize = 1024;
const char Buffer::kCRLF[] = "\r\n";

Buffer::Buffer(size_t initialSize)
    : buffer_(kCheapPrepend + initialSize)
    , readerIndex_(kCheapPrepend)
    , writerIndex_(kCheapPrepend)
{
    Q_ASSERT(readableBytes() == 0);
    Q_ASSERT(writableBytes() == initialSize);
    Q_ASSERT(prependableBytes() == kCheapPrepend);
}

Buffer::Buffer(const char *data, size_t len)
    : buffer_(kCheapPrepend + len)
    , readerIndex_(kCheapPrepend)
    , writerIndex_(kCheapPrepend + len)
{
    Q_ASSERT(data != nullptr && len > 0);
    memcpy(&*buffer_.begin() + kCheapPrepend, data, len);
    Q_ASSERT(readableBytes() == len);
    Q_ASSERT(writableBytes() == 0);
    Q_ASSERT(prependableBytes() == kCheapPrepend);
}

Buffer::Buffer(Buffer &&buffer)
    : buffer_(std::move(buffer.buffer_))
    , readerIndex_(buffer.readerIndex_)
    , writerIndex_(buffer.writerIndex_)
{

}
