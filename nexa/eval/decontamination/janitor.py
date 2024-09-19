import pickle
import re
import string
import traceback
from typing import Iterator, List, Sequence, Tuple, TypeVar


# This is a cpp module. Compile janitor_util.cpp with:
# c++ -O3 -Wall -shared -std=c++11 -fPIC $(python3 -m pybind11 --includes) janitor_util.cpp -o janitor_util$(python3-config --extension-suffix) -undefined dynamic_lookup
try:
    import janitor_util

    JANITOR_CPP = True
except Exception:
    print("WARNING: C++ module could not be loaded. Janitor running in python mode")
    traceback.print_exc()
    JANITOR_CPP = False

T = TypeVar("T")


# Implementation from nltk source
# https://www.nltk.org/_modules/nltk/util.html
def form_ngrams(sequence: Iterator[T], n: int) -> Iterator[Tuple[T, ...]]:
    history = []
    while n > 1:
        # PEP 479, prevent RuntimeError from being raised when StopIteration bubbles out of generator
        try:
            next_item = next(sequence)
        except StopIteration:
            # no more data, terminate the generator
            return
        history.append(next_item)
        n -= 1
    for item in sequence:
        history.append(item)
        yield tuple(history)
        del history[0]


def word_ngrams(s: str, n: int) -> Iterator[str]:
    """Splits a string into ngram words"""
    tokens = s.split()  # not a generator :(
    ngram_seqs = form_ngrams(iter(tokens), n)
    return (" ".join(ngram) for ngram in ngram_seqs)


# Does character sequences only - combined faster function to play around with later
# def word_ngrams_indices_combined(sequence, n):
#     current_word = ""
#     history = []
#     gap = False;
#     start = 0
#     end = 0
#     for character in sequence:
#         if character == " ":
#             if not gap:
#                 gap = True
#                 history.append(current_word)
#                 end += len(current_word) - 1
#                 current_word = ""
#                 if len(history) == n:
#                     yield (tuple(history), start, end)
#                     del history[0]
#                     start = end + 1
#                     end = start
#         else:
#             gap = False
#             current_word += character


# https://stackoverflow.com/questions/13734451/string-split-with-indices-in-python
def split_indices(s: str) -> Iterator[Tuple[str, Tuple[int, int]]]:
    """Splits a string on whitespaces and records the indices of each in the original string.
    @:return generator((word, (start_idx, end_idx)), ...)
    """
    return ((m.group(0), (m.start(), m.end() - 1)) for m in re.finditer(r"\S+", s))


def word_ngrams_indices(s: str, n: int) -> Iterator[Tuple[str, Tuple[int, int]]]:
    """Splits a string into pairs of (ngram words, their start/end indices)"""
    tokens_with_indices = split_indices(s)

    # Generator of ngrams of (word, idx_pairs)
    # (
    #   [(word, (start,end)), (word, (start, end))...],
    #   [(word, (start, end)), ...],
    #   ...
    # )
    ngram_seqs_with_indices = form_ngrams(tokens_with_indices, n)

    # Generator of pairs of word and index ngrams
    # (
    #   ([word, word, ...], [(start,end), (start,end), ...]),
    #   ...
    # )
    ngram_indices_pairs = (
        zip(*ngram_with_indices) for ngram_with_indices in ngram_seqs_with_indices
    )

    # Generator of ( (word_ngram, (start, end)), (word_ngram, start, end)), ...)
    return (
        (" ".join(ngram_seq), (indices[0][0], indices[-1][1]))
        for ngram_seq, indices in ngram_indices_pairs
    )


class Janitor:
    # FIXME delete_chars: Should anything else go here? Special chars?
    def __init__(
        self,
        ngram_n: int = 13,
        window_to_remove: int = 200,
        too_dirty_cutoff: int = 10,
        minimum_slice_length: int = 200,
        delete_chars: str = string.punctuation,
    ) -> None:
        self.ngram_n = ngram_n
        self.window_to_remove = window_to_remove
        self.too_dirty_cutoff = too_dirty_cutoff
        self.minimum_slice_length = minimum_slice_length
        self.delete_chars = delete_chars

        self.dirt_ngrams = set()

        # If in python, we'll translate uppercase to lowercase and delete naughty characters.
        # This is fast by python standards
        # https://stackoverflow.com/questions/638893/what-is-the-most-efficient-way-in-python-to-convert-a-string-to-all-lowercase-st
        self.translation_table = str.maketrans(
            string.ascii_lowercase + string.ascii_uppercase,  # These characters
            string.ascii_lowercase * 2,  # Become these characters
            self.delete_chars,  # These are deleted
        )

    ##############
    # I/O for saving contamination ngrams
    ##############

    def save_contamination_ngrams(self, filename: str) -> None:
        with open(filename, "wb") as fp:
            pickle.dump(filename, fp)

    def load_contamination_ngrams(self, filename: str) -> None:
        with open(filename, "rb") as fp:
            self.dirt_ngrams = pickle.load(fp)

    ##############
    # Call these :)
    ##############

    def register_contaminant(self, dirt_string: str) -> None:
        """Register a string as contamination to be removed, e.g. a test set
        This breaks the dirt_string into ngrams to store for future cleaning"""
        if JANITOR_CPP:
            return self.register_contaminant_cpp(dirt_string)
        else:
            print("WARNING: Janitor running in python mode")
            return self.register_contaminant_python(dirt_string)

    def clean(self, dirty_string: str) -> List[str]:
        """Clean a string (e.g. a training set) by removing all ngrams previously
        registered as contaminants. Returns a list of clean chunks, or empty if
        the string was too dirty"""
        if JANITOR_CPP:
            return self.clean_cpp(dirty_string)
        else:
            print("WARNING: Janitor running in python mode")
            return self.clean_python(dirty_string)

    def _split_chunks(
        self, dirty_string: str, dirty_parts: Sequence[Tuple]
    ) -> List[str]:
        clean_chunks = []
        splice_idx = 0
        end = -1
        for i, (ngram, start, end) in enumerate(dirty_parts):
            if i >= self.too_dirty_cutoff:
                return []
            start = max(0, start - self.window_to_remove)
            end = min(len(dirty_string), end + self.window_to_remove)

            if start - splice_idx > self.minimum_slice_length:
                clean_chunks.append(dirty_string[splice_idx:start])
            splice_idx = end

        if end < len(dirty_string) - self.minimum_slice_length:
            clean_chunks.append(dirty_string[end + 1 :])

        return clean_chunks

    ##############
    # Fast C++
    ##############

    def register_contaminant_cpp(self, dirt_string) -> None:
        self.dirt_ngrams.update(
            janitor_util.clean_ngram(dirt_string, self.delete_chars, self.ngram_n)
        )

    def clean_cpp(self, dirty_string: str) -> List[str]:
        contamination_indices = janitor_util.clean_ngram_with_indices(
            dirty_string, self.delete_chars, self.ngram_n
        )
        return self._split_chunks(dirty_string, contamination_indices)

    ##############
    # Slow python
    ##############

    def normalize_string(self, s: str) -> str:
        return s.translate(self.translation_table)

    def register_contaminant_python(self, dirt_string: str) -> None:
        self.dirt_ngrams.update(
            word_ngrams(self.normalize_string(dirt_string), self.ngram_n)
        )

    def clean_python(self, dirty_string: str) -> List[str]:
        contamination_indices = (
            (None, *idx_pair)
            for dirty_ngram, idx_pair in word_ngrams_indices(dirty_string, self.ngram_n)
            if self.normalize_string(dirty_ngram) in self.dirt_ngrams
        )
        return self._split_chunks(dirty_string, contamination_indices)