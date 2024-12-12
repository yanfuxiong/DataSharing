#pragma once
#include <QtEndian>
#include <QByteArray>

#include <algorithm>
#include <vector>
#include <string>
#include <cstring>

class Buffer
{
public:
    explicit Buffer(size_t initialSize = kInitialSize);
    Buffer(const char *data, size_t len);
    Buffer(Buffer &&buffer);
    Buffer(const Buffer &) = default;
    Buffer &operator = (const Buffer &) = default;
    Buffer &operator = (Buffer &&buffer)
    {
        if (this == &buffer) {
            return *this;
        }
        buffer_ = std::move(buffer.buffer_);
        readerIndex_ = buffer.readerIndex_;
        writerIndex_ = buffer.writerIndex_;
        return *this;
    }
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
        Q_ASSERT(peek() <= start);
        Q_ASSERT(start <= beginWrite());
        const char *crlf = std::search(start, beginWrite(), kCRLF, kCRLF + 2);
        return crlf == beginWrite() ? nullptr : crlf;
    }

    const char *findEOL() const
    {
        const void *eol = memchr(peek(), '\n', readableBytes());
        return static_cast<const char*>(eol);
    }

    const char *findEOL(const char *start) const
    {
        Q_ASSERT(peek() <= start);
        Q_ASSERT(start <= beginWrite());
        const void *eol = memchr(start, '\n', static_cast<size_t>(beginWrite() - start));
        return static_cast<const char*>(eol);
    }

    void retrieve(size_t len)
    {
        Q_ASSERT(len <= readableBytes());
        if (len < readableBytes()) {
            readerIndex_ += len;
        } else {
            retrieveAll();
        }
    }

    void retrieveUntil(const char *end)
    {
        Q_ASSERT(peek() <= end);
        Q_ASSERT(end <= beginWrite());
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

    QByteArray retrieveAllAsByteArray()
    {
        return retrieveAsByteArray(readableBytes());
    }

    std::string retrieveAsString(size_t len)
    {
        Q_ASSERT(len <= readableBytes());
        std::string result(peek(), len);
        retrieve(len);
        return result;
    }

    QByteArray retrieveAsByteArray(size_t len)
    {
        Q_ASSERT(len <= readableBytes());
        QByteArray result(peek(), static_cast<int>(len));
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

    void append(const QByteArray &data)
    {
        append(data.constData(), static_cast<size_t>(data.length()));
    }

    void ensureWritableBytes(size_t len)
    {
        if (writableBytes() < len) {
            makeSpace(len);
        }
        Q_ASSERT(writableBytes() >= len);
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
        Q_ASSERT(len <= writableBytes());
        writerIndex_ += len;
    }

    void unwrite(size_t len)
    {
        Q_ASSERT(len <= readableBytes());
        writerIndex_ -= len;
    }

    void appendUInt64(uint64_t x)
    {
        // Linux 64位环境下：uint64_t 为 unsigned long int, quint64 为unsigned long long,
        // 使用qToBigEndian时如果使用uint64_t 会报错
        uint64_t be64 = qToBigEndian<quint64>(static_cast<quint64>(x));
        append(&be64, sizeof(be64));
    }

    void appendUInt32(uint32_t x)
    {
        uint32_t be32 = qToBigEndian<uint32_t>(x);
        append(&be32, sizeof(be32));
    }

    void appendUInt16(uint16_t x)
    {
        uint16_t be16 = qToBigEndian<uint16_t>(x);
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
        Q_ASSERT(sizeof(uint64_t) <= readableBytes());
        uint64_t be64 = 0;
        memcpy(&be64, peek(), sizeof(be64));
        return qFromBigEndian<quint64>(static_cast<quint64>(be64));
    }

    uint32_t peekUInt32() const
    {
        Q_ASSERT(sizeof(uint32_t) <= readableBytes());
        uint32_t be32 = 0;
        memcpy(&be32, peek(), sizeof(be32));
        return qFromBigEndian<uint32_t>(be32);
    }

    uint16_t peekUInt16() const
    {
        Q_ASSERT(sizeof(uint16_t) <= readableBytes());
        uint16_t be16 = 0;
        memcpy(&be16, peek(), sizeof(be16));
        return qFromBigEndian<uint16_t>(be16);
    }

    void prependUInt64(uint64_t x)
    {
        uint64_t be64 = qToBigEndian<quint64>(static_cast<quint64>(x));
        prepend(&be64, sizeof(be64));
    }

    void prependUInt32(uint32_t x)
    {
        uint32_t be32 = qToBigEndian<uint32_t>(x);
        prepend(&be32, sizeof(be32));
    }

    void prependUInt16(uint16_t x)
    {
        uint16_t be16 = qToBigEndian<uint16_t>(x);
        prepend(&be16, sizeof(be16));
    }

    void prepend(const void *data, size_t len)
    {
        Q_ASSERT(len <= prependableBytes());
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
        Q_ASSERT(len > writableBytes());
        if (writableBytes() + prependableBytes() < len + kCheapPrepend) {
            buffer_.resize(writerIndex_+len);
        } else {
            Q_ASSERT(kCheapPrepend < readerIndex_);
            size_t readable = readableBytes();
            memcpy(begin() + kCheapPrepend, begin() + readerIndex_, readable);
            readerIndex_ = kCheapPrepend;
            writerIndex_ = readerIndex_ + readable;
            Q_ASSERT(readable == readableBytes());
        }
    }

private:
    std::vector<char> buffer_;
    size_t readerIndex_;
    size_t writerIndex_;

    static const size_t kCheapPrepend;
    static const size_t kInitialSize;
    static const char kCRLF[];
};
