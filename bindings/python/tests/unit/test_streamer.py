# Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
# SPDX-License-Identifier: BSD-3-Clause

from __future__ import annotations

from geniex.generation.streamer import TextIteratorStreamer


def test_callback_enqueues_when_not_cancelled():
    streamer = TextIteratorStreamer()
    callback = streamer._make_callback()

    assert callback(b'hello', None) is True
    assert streamer._queue.get_nowait() == 'hello'


def test_cancel_drops_rejected_token_keeps_accepted():
    streamer = TextIteratorStreamer()
    callback = streamer._make_callback()

    assert callback(b' first', None) is True
    streamer.cancel()
    assert callback(b' rejected', None) is False

    assert streamer._queue.get_nowait() == ' first'
    assert streamer._queue.empty()
