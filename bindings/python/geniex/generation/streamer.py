# Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
# SPDX-License-Identifier: BSD-3-Clause

from __future__ import annotations

import queue
import threading
from ctypes import c_void_p
from typing import Callable, Iterator

from .._ffi._types import geniex_token_callback
from .output import GenerateOutput

_SENTINEL = object()


class TextIteratorStreamer:
    """Yields decoded text chunks as the model generates them.

    Usage::

        for chunk in model.generate(prompt, stream=True):
            print(chunk, end='', flush=True)
        # After the loop: streamer.output holds the final GenerateOutput
    """

    def __init__(self) -> None:
        self._queue: queue.Queue = queue.Queue()
        self._output: GenerateOutput | None = None
        self._error: BaseException | None = None
        self._thread: threading.Thread | None = None
        self._cancelled = threading.Event()

    @property
    def output(self) -> GenerateOutput | None:
        return self._output

    def cancel(self) -> None:
        """Stop generation at the next token boundary."""
        self._cancelled.set()

    def _make_callback(self) -> geniex_token_callback:
        q = self._queue
        cancelled = self._cancelled

        @geniex_token_callback
        def _cb(token_bytes: bytes, _userdata: c_void_p) -> bool:
            if cancelled.is_set():
                # A callback that returns False stops generation, but the token
                # it rejected is still excluded from GenerateOutput.text. Drop
                # it here so the stream never yields text that is missing from
                # the final output.
                return False
            if token_bytes is not None:
                q.put(token_bytes.decode('utf-8', errors='replace'))
            # Returning True accepts the token into both the KV cache and
            # GenerateOutput.text. If cancel() races with this callback, finish
            # accepting the token rather than streaming one the SDK discards;
            # the next callback observes the flag and stops.
            return True

        # Retain a reference so GC doesn't free the callback while C holds the pointer.
        self._cb_ref = _cb
        return _cb

    def _run(self, generate_fn: Callable[[], GenerateOutput]) -> None:
        try:
            result = generate_fn()
            self._output = result
        except BaseException as exc:
            self._error = exc
        finally:
            self._queue.put(_SENTINEL)

    def start(self, generate_fn: Callable[[], GenerateOutput]) -> None:
        self._thread = threading.Thread(target=self._run, args=(generate_fn,), daemon=True)
        self._thread.start()

    def __iter__(self) -> Iterator[str]:
        while True:
            item = self._queue.get()
            if item is _SENTINEL:
                break
            yield item
        if self._error is not None:
            raise self._error
        if self._thread is not None:
            self._thread.join()
