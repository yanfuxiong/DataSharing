#pragma once
#include "global.h"

namespace sunkang {

class Buffer
{
public:
    explicit Buffer(size_t initialSize = kInitialSize);
    Buffer(const char *data, size_t len);
    Buffer(Buffer &&buffer);
    Buffer(const Buffer &) = default;
    Buffer &operator = (Buffer &&buffer);
    Buffer &operator = (const Buffer &) = default;
    ~Buffer() = default;

    void swap(Buffer &rhs)
    {
        buffer_.swap(rhs.buffer_);
        std::swap(readerIndex_, rhs.readerIndex_);
        std::swap(writerIndex_, rhs.writerIndex_);
    }

    size_t readableBytes() const
    {
        return writerIndex_ - readerIndex_;
    }

    size_t writableBytes() const
    {
        return buffer_.size() - writerIndex_;
    }

    size_t prependableBytes() const
    {
        return readerIndex_;
    }

    const char *peek() const
    {
        return begin() + readerIndex_;
    }

    const char *findCRLF() const
    {
        const char *crlf = std::search(peek(), beginWrite(), kCRLF, kCRLF + 2);
        return crlf == beginWrite() ? nullptr : crlf;
    }

    const char *findCRLF(const char *start) const
    {
        assert(peek() <= start);
        assert(start <= beginWrite());
        const char *crlf = std::search(start, beginWrite(), kCRLF, kCRLF+2);
        return crlf == beginWrite() ? nullptr : crlf;
    }

    const char *findEOL() const
    {
        const void *eol = memchr(peek(), '\n', readableBytes());
        return static_cast<const char*>(eol);
    }

    const char *findEOL(const char *start) const
    {
        assert(peek() <= start);
        assert(start <= beginWrite());
        const void *eol = memchr(start, '\n', static_cast<size_t>(beginWrite() - start));
        return static_cast<const char*>(eol);
    }

    void retrieve(size_t len)
    {
        assert(len <= readableBytes());
        if (len < readableBytes()) {
            readerIndex_ += len;
        } else {
            retrieveAll();
        }
    }

    void retrieveUntil(const char *end)
    {
        assert(peek() <= end);
        assert(end <= beginWrite());
        retrieve(static_cast<size_t>(end - peek()));
    }
    void retrieveUInt64()
    {
        retrieve(sizeof(uint64_t));
    }

    void retrieveUInt32()
    {
        retrieve(sizeof(uint32_t));
    }

    void retrieveUInt16()
    {
        retrieve(sizeof(uint16_t));
    }

    void retrieveAll()
    {
        readerIndex_ = kCheapPrepend;
        writerIndex_ = kCheapPrepend;
    }

    std::string retrieveAllAsString()
    {
        return retrieveAsString(readableBytes());
    }

    std::string retrieveAsString(size_t len)
    {
        assert(len <= readableBytes());
        std::string result(peek(), len);
        retrieve(len);
        return result;
    }

    void append(const char *data, size_t len)
    {
        ensureWritableBytes(len);
        memcpy(beginWrite(), data, len);
        hasWritten(len);
    }

    void append(const void *data, size_t len)
    {
        append(static_cast<const char *>(data), len);
    }

    void append(const std::string &data)
    {
        append(data.data(), data.size());
    }

    void append(const std::wstring &data)
    {
        append(data.data(), data.size() * sizeof (wchar_t));
    }

    void ensureWritableBytes(size_t len)
    {
        if (writableBytes() < len) {
            makeSpace(len);
        }
        assert(writableBytes() >= len);
    }

    char *beginWrite()
    {
        return begin() + writerIndex_;
    }

    const char *beginWrite() const
    {
        return begin() + writerIndex_;
    }

    void hasWritten(size_t len)
    {
        assert(len <= writableBytes());
        writerIndex_ += len;
    }

    void unwrite(size_t len)
    {
        assert(len <= readableBytes());
        writerIndex_ -= len;
    }

    void appendUInt64(uint64_t x)
    {
        uint64_t be64 = bytes_swap_64(x);
        append(&be64, sizeof(be64));
    }

    void appendUInt32(uint32_t x)
    {
        uint32_t be32 = asio::detail::socket_ops::host_to_network_long(x);
        append(&be32, sizeof(be32));
    }

    void appendUInt16(uint16_t x)
    {
        uint16_t be16 = asio::detail::socket_ops::host_to_network_short(x);
        append(&be16, sizeof(be16));
    }

    uint64_t readUInt64()
    {
        uint64_t result = peekUInt64();
        retrieveUInt64();
        return result;
    }

    uint32_t readUInt32()
    {
        uint32_t result = peekUInt32();
        retrieveUInt32();
        return result;
    }

    uint16_t readUInt16()
    {
        uint16_t result = peekUInt16();
        retrieveUInt16();
        return result;
    }

    uint64_t peekUInt64() const
    {
        assert(sizeof(uint64_t) <= readableBytes());
        uint64_t be64 = 0;
        memcpy(&be64, peek(), sizeof(be64));
        return bytes_swap_64(be64);
    }

    uint32_t peekUInt32() const
    {
        assert(sizeof(uint32_t) <= readableBytes());
        uint32_t be32 = 0;
        memcpy(&be32, peek(), sizeof(be32));
        return asio::detail::socket_ops::network_to_host_long(be32);
    }

    uint16_t peekUInt16() const
    {
        assert(sizeof(uint16_t) <= readableBytes());
        uint16_t be16 = 0;
        memcpy(&be16, peek(), sizeof(be16));
        return asio::detail::socket_ops::network_to_host_short(be16);
    }

    void prependUInt64(uint64_t x)
    {
        uint64_t be64 = bytes_swap_64(x);
        prepend(&be64, sizeof(be64));
    }

    void prependUInt32(uint32_t x)
    {
        uint32_t be32 = asio::detail::socket_ops::host_to_network_long(x);
        prepend(&be32, sizeof(be32));
    }

    void prependUInt16(uint16_t x)
    {
        uint16_t be16 = asio::detail::socket_ops::host_to_network_short(x);
        prepend(&be16, sizeof(be16));
    }

    void prepend(const void *data, size_t len)
    {
        assert(len <= prependableBytes());
        readerIndex_ -= len;
        memcpy(begin() + readerIndex_, data, len);
    }

    size_t internalCapacity() const
    {
        return buffer_.capacity();
    }

private:
    char *begin()
    {
        return &*buffer_.begin();
    }

    const char *begin() const
    {
        return &*buffer_.begin();
    }

    void makeSpace(size_t len)
    {
        assert(len > writableBytes());
        if (writableBytes() + prependableBytes() < len + kCheapPrepend) {
            buffer_.resize(writerIndex_+len);
        } else {
            assert(kCheapPrepend < readerIndex_);
            size_t readable = readableBytes();
            memcpy(begin() + kCheapPrepend, begin() + readerIndex_, readable);
            readerIndex_ = kCheapPrepend;
            writerIndex_ = readerIndex_ + readable;
            assert(readable == readableBytes());
        }
    }

    static inline uint64_t bytes_swap_64(uint64_t source)
    {
        return 0
            | ((source & 0x00000000000000ffull) << 56)
            | ((source & 0x000000000000ff00ull) << 40)
            | ((source & 0x0000000000ff0000ull) << 24)
            | ((source & 0x00000000ff000000ull) << 8)
            | ((source & 0x000000ff00000000ull) >> 8)
            | ((source & 0x0000ff0000000000ull) >> 24)
            | ((source & 0x00ff000000000000ull) >> 40)
            | ((source & 0xff00000000000000ull) >> 56);
    }

private:
    std::vector<char> buffer_;
    size_t readerIndex_;
    size_t writerIndex_;

    static const size_t kCheapPrepend;
    static const size_t kInitialSize;
    static const char kCRLF[];
};

} // namespace sunkang
