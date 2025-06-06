//
// ssl/detail/stream_core.hpp
// ~~~~~~~~~~~~~~~~~~~~~~~~~~
//
// Copyright (c) 2003-2024 Christopher M. Kohlhoff (chris at kohlhoff dot com)
//
// Distributed under the Boost Software License, Version 1.0. (See accompanying
// file LICENSE_1_0.txt or copy at http://www.boost.org/LICENSE_1_0.txt)
//

#ifndef ASIO_SSL_DETAIL_STREAM_CORE_HPP
#define ASIO_SSL_DETAIL_STREAM_CORE_HPP

#if defined(_MSC_VER) && (_MSC_VER >= 1200)
# pragma once
#endif // defined(_MSC_VER) && (_MSC_VER >= 1200)

#include "asio/detail/config.hpp"

#if defined(ASIO_HAS_BOOST_DATE_TIME)
# include "asio/deadline_timer.hpp"
#else // defined(ASIO_HAS_BOOST_DATE_TIME)
# include "asio/steady_timer.hpp"
#endif // defined(ASIO_HAS_BOOST_DATE_TIME)
#include "asio/ssl/detail/engine.hpp"
#include "asio/buffer.hpp"

#include "asio/detail/push_options.hpp"

namespace asio {
namespace ssl {
namespace detail {

struct stream_core
{
  // According to the OpenSSL documentation, this is the buffer size that is
  // sufficient to hold the largest possible TLS record.
  enum { max_tls_record_size = 17 * 1024 };

  template <typename Executor>
  stream_core(SSL_CTX* context, const Executor& ex)
    : engine_(context),
      pending_read_(ex),
      pending_write_(ex),
      output_buffer_space_(max_tls_record_size),
      output_buffer_(asio::buffer(output_buffer_space_)),
      input_buffer_space_(max_tls_record_size),
      input_buffer_(asio::buffer(input_buffer_space_))
  {
    pending_read_.expires_at(neg_infin());
    pending_write_.expires_at(neg_infin());
  }

  template <typename Executor>
  stream_core(SSL* ssl_impl, const Executor& ex)
    : engine_(ssl_impl),
      pending_read_(ex),
      pending_write_(ex),
      output_buffer_space_(max_tls_record_size),
      output_buffer_(asio::buffer(output_buffer_space_)),
      input_buffer_space_(max_tls_record_size),
      input_buffer_(asio::buffer(input_buffer_space_))
  {
    pending_read_.expires_at(neg_infin());
    pending_write_.expires_at(neg_infin());
  }

  stream_core(stream_core&& other)
    : engine_(static_cast<engine&&>(other.engine_)),
#if defined(ASIO_HAS_BOOST_DATE_TIME)
      pending_read_(
         static_cast<asio::deadline_timer&&>(
           other.pending_read_)),
      pending_write_(
         static_cast<asio::deadline_timer&&>(
           other.pending_write_)),
#else // defined(ASIO_HAS_BOOST_DATE_TIME)
      pending_read_(
         static_cast<asio::steady_timer&&>(
           other.pending_read_)),
      pending_write_(
         static_cast<asio::steady_timer&&>(
           other.pending_write_)),
#endif // defined(ASIO_HAS_BOOST_DATE_TIME)
      output_buffer_space_(
          static_cast<std::vector<unsigned char>&&>(
            other.output_buffer_space_)),
      output_buffer_(other.output_buffer_),
      input_buffer_space_(
          static_cast<std::vector<unsigned char>&&>(
            other.input_buffer_space_)),
      input_buffer_(other.input_buffer_),
      input_(other.input_)
  {
    other.output_buffer_ = asio::mutable_buffer(0, 0);
    other.input_buffer_ = asio::mutable_buffer(0, 0);
    other.input_ = asio::const_buffer(0, 0);
  }

  ~stream_core()
  {
  }

  stream_core& operator=(stream_core&& other)
  {
    if (this != &other)
    {
      engine_ = static_cast<engine&&>(other.engine_);
#if defined(ASIO_HAS_BOOST_DATE_TIME)
      pending_read_ =
        static_cast<asio::deadline_timer&&>(
          other.pending_read_);
      pending_write_ =
        static_cast<asio::deadline_timer&&>(
          other.pending_write_);
#else // defined(ASIO_HAS_BOOST_DATE_TIME)
      pending_read_ =
        static_cast<asio::steady_timer&&>(
          other.pending_read_);
      pending_write_ =
        static_cast<asio::steady_timer&&>(
          other.pending_write_);
#endif // defined(ASIO_HAS_BOOST_DATE_TIME)
      output_buffer_space_ =
        static_cast<std::vector<unsigned char>&&>(
          other.output_buffer_space_);
      output_buffer_ = other.output_buffer_;
      input_buffer_space_ =
        static_cast<std::vector<unsigned char>&&>(
          other.input_buffer_space_);
      input_buffer_ = other.input_buffer_;
      input_ = other.input_;
      other.output_buffer_ = asio::mutable_buffer(0, 0);
      other.input_buffer_ = asio::mutable_buffer(0, 0);
      other.input_ = asio::const_buffer(0, 0);
    }
    return *this;
  }

  // The SSL engine.
  engine engine_;

#if defined(ASIO_HAS_BOOST_DATE_TIME)
  // Timer used for storing queued read operations.
  asio::deadline_timer pending_read_;

  // Timer used for storing queued write operations.
  asio::deadline_timer pending_write_;

  // Helper function for obtaining a time value that always fires.
  static asio::deadline_timer::time_type neg_infin()
  {
    return boost::posix_time::neg_infin;
  }

  // Helper function for obtaining a time value that never fires.
  static asio::deadline_timer::time_type pos_infin()
  {
    return boost::posix_time::pos_infin;
  }

  // Helper function to get a timer's expiry time.
  static asio::deadline_timer::time_type expiry(
      const asio::deadline_timer& timer)
  {
    return timer.expires_at();
  }
#else // defined(ASIO_HAS_BOOST_DATE_TIME)
  // Timer used for storing queued read operations.
  asio::steady_timer pending_read_;

  // Timer used for storing queued write operations.
  asio::steady_timer pending_write_;

  // Helper function for obtaining a time value that always fires.
  static asio::steady_timer::time_point neg_infin()
  {
    return (asio::steady_timer::time_point::min)();
  }

  // Helper function for obtaining a time value that never fires.
  static asio::steady_timer::time_point pos_infin()
  {
    return (asio::steady_timer::time_point::max)();
  }

  // Helper function to get a timer's expiry time.
  static asio::steady_timer::time_point expiry(
      const asio::steady_timer& timer)
  {
    return timer.expiry();
  }
#endif // defined(ASIO_HAS_BOOST_DATE_TIME)

  // Buffer space used to prepare output intended for the transport.
  std::vector<unsigned char> output_buffer_space_;

  // A buffer that may be used to prepare output intended for the transport.
  asio::mutable_buffer output_buffer_;

  // Buffer space used to read input intended for the engine.
  std::vector<unsigned char> input_buffer_space_;

  // A buffer that may be used to read input intended for the engine.
  asio::mutable_buffer input_buffer_;

  // The buffer pointing to the engine's unconsumed input.
  asio::const_buffer input_;
};

} // namespace detail
} // namespace ssl
} // namespace asio

#include "asio/detail/pop_options.hpp"

#endif // ASIO_SSL_DETAIL_STREAM_CORE_HPP
