#include "buffer_x.h"

namespace sunkang {

const size_t Buffer::kCheapPrepend = 8;
const size_t Buffer::kInitialSize = 1024;
const char Buffer::kCRLF[] = "\r\n";

Buffer::Buffer(size_t initialSize)
    : buffer_(kCheapPrepend + initialSize)
    , readerIndex_(kCheapPrepend)
    , writerIndex_(kCheapPrepend)
{
    assert(readableBytes() == 0);
    assert(writableBytes() == initialSize);
    assert(prependableBytes() == kCheapPrepend);
}

Buffer::Buffer(const char *data, size_t len)
    : buffer_(kCheapPrepend + len)
    , readerIndex_(kCheapPrepend)
    , writerIndex_(kCheapPrepend + len)
{
    assert(data != nullptr && len > 0);
    memcpy(&*buffer_.begin() + kCheapPrepend, data, len);
    assert(readableBytes() == len);
    assert(writableBytes() == 0);
    assert(prependableBytes() == kCheapPrepend);
}

Buffer::Buffer(Buffer &&buffer)
    : buffer_(std::move(buffer.buffer_))
    , readerIndex_(buffer.readerIndex_)
    , writerIndex_(buffer.writerIndex_)
{

}

Buffer &Buffer::operator = (Buffer &&buffer)
{
    if (this == &buffer) {
        return *this;
    }
    buffer_ = std::move(buffer.buffer_);
    readerIndex_ = buffer.readerIndex_;
    writerIndex_ = buffer.writerIndex_;
    assert(readableBytes() == writerIndex_ - readerIndex_);
    return *this;
}

} // namespace sunkang
