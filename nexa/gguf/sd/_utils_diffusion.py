import os
import sys

# Avoid "LookupError: unknown encoding: ascii" when open() called in a destructor
outnull_file = open(os.devnull, "w")
errnull_file = open(os.devnull, "w")

STDOUT_FILENO = 1
STDERR_FILENO = 2


class suppress_stdout_stderr(object):
    """
    Stops all output to stdout and stderr when used as a context manager (GGML will otherwise still print logs).
    Source: https://github.com/abetlen/llama-cpp-python/blob/main/llama_cpp/_utils.py
    """

    # NOTE: these must be "saved" here to avoid exceptions when using
    #       this context manager inside of a __del__ method
    sys = sys
    os = os

    def __init__(self, disable: bool = True):
        self.disable = disable

    # Oddly enough this works better than the contextlib version
    def __enter__(self):
        if self.disable:
            return self

        self.old_stdout_fileno_undup = STDOUT_FILENO
        self.old_stderr_fileno_undup = STDERR_FILENO

        self.old_stdout_fileno = self.os.dup(self.old_stdout_fileno_undup)
        self.old_stderr_fileno = self.os.dup(self.old_stderr_fileno_undup)

        self.old_stdout = self.sys.stdout
        self.old_stderr = self.sys.stderr

        self.os.dup2(outnull_file.fileno(), self.old_stdout_fileno_undup)
        self.os.dup2(errnull_file.fileno(), self.old_stderr_fileno_undup)

        self.sys.stdout = outnull_file
        self.sys.stderr = errnull_file
        return self

    def __exit__(self, *_):
        if self.disable:
            return

        # Check if sys.stdout and sys.stderr have fileno method
        self.sys.stdout = self.old_stdout
        self.sys.stderr = self.old_stderr

        self.os.dup2(self.old_stdout_fileno, self.old_stdout_fileno_undup)
        self.os.dup2(self.old_stderr_fileno, self.old_stderr_fileno_undup)

        self.os.close(self.old_stdout_fileno)
        self.os.close(self.old_stderr_fileno)
