# Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
# SPDX-License-Identifier: BSD-3-Clause

"""Cancellation semantics for TextIteratorStreamer.

The SDK stops generation when a token callback returns False, but the token it
rejected is not part of GenerateOutput.text. The streamer must therefore not
deliver that token to the consumer: otherwise the concatenated stream contains
text that the final output does not.
"""

from __future__ import annotations

from geniex.generation.streamer import TextIteratorStreamer


def test_callback_enqueues_when_not_cancelled():
    streamer = TextIteratorStreamer()
    callback = streamer._make_callback()

    assert callback(b'hello', None) is True
    assert streamer._queue.get_nowait() == 'hello'


def test_cancel_before_callback_drops_the_token():
    streamer = TextIteratorStreamer()
    callback = streamer._make_callback()
    streamer.cancel()

    assert callback(b'late', None) is False
    assert streamer._queue.empty()


def test_cancel_mid_stream_keeps_accepted_tokens_only():
    streamer = TextIteratorStreamer()
    callback = streamer._make_callback()

    assert callback(b' first', None) is True
    streamer.cancel()
    assert callback(b' rejected', None) is False

    assert streamer._queue.get_nowait() == ' first'
    assert streamer._queue.empty()
