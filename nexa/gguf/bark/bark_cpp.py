# auto-generated file
import ctypes
from nexa.gguf.lib_utils import load_library
from typing import Pointer


# Specify the base name of the shared library to load
_lib_base_name = "bark"

# Load the library
_lib = load_library(_lib_base_name)



bark_context_p = ctypes.c_void_p

bark_model_p = ctypes.c_void_p

bark_vocab_p = ctypes.c_void_p

gpt_model_p = ctypes.c_void_p

class bark_statistics(ctypes.Structure):
    _fields_ = [
        ("t_load_us", ctypes.c_int64),
        ("t_eval_us", ctypes.c_int64),
        ("t_semantic_us", ctypes.c_int64),
        ("t_coarse_us", ctypes.c_int64),
        ("t_fine_us", ctypes.c_int64),
        ("n_sample_semantic", ctypes.c_void_p),
        ("n_sample_coarse", ctypes.c_void_p),
        ("n_sample_fine", ctypes.c_void_p),
    ]

class bark_context_params(ctypes.Structure):
    _fields_ = [
        ("verbosity", ctypes.c_int),
        ("temp", ctypes.c_float),
        ("fine_temp", ctypes.c_float),
        ("min_eos_p", ctypes.c_float),
        ("sliding_window_size", ctypes.c_void_p),
        ("max_coarse_history", ctypes.c_void_p),
        ("sample_rate", ctypes.c_void_p),
        ("target_bandwidth", ctypes.c_void_p),
        ("cls_token_id", ctypes.c_void_p),
        ("sep_token_id", ctypes.c_void_p),
        ("n_steps_text_encoder", ctypes.c_void_p),
        ("text_pad_token", ctypes.c_void_p),
        ("text_encoding_offset", ctypes.c_void_p),
        ("semantic_rate_hz", ctypes.c_float),
        ("semantic_pad_token", ctypes.c_void_p),
        ("semantic_vocab_size", ctypes.c_void_p),
        ("semantic_infer_token", ctypes.c_void_p),
        ("coarse_rate_hz", ctypes.c_float),
        ("coarse_infer_token", ctypes.c_void_p),
        ("coarse_semantic_pad_token", ctypes.c_void_p),
        ("n_coarse_codebooks", ctypes.c_void_p),
        ("n_fine_codebooks", ctypes.c_void_p),
        ("codebook_size", ctypes.c_void_p),
        ("progress_callback", ctypes.c_void_p),
        ("progress_callback_user_data", ctypes.c_void_p),
    ]

def bark_context_default_params() -> bark_context_params:
    return _lib.bark_context_default_params()

_lib.bark_context_default_params.argtypes = []
_lib.bark_context_default_params.restype = bark_context_params


def bark_load_model(model_path: ctypes.c_char_p,
    params: bark_context_params,
    seed: ctypes.c_void_p) -> bark_context_p:
    return _lib.bark_load_model(model_path, params, seed)

_lib.bark_load_model.argtypes = [ctypes.c_char_p, bark_context_params, ctypes.c_void_p]
_lib.bark_load_model.restype = bark_context_p


def bark_generate_audio(bctx: bark_context_p,
    text: ctypes.c_char_p,
    n_threads: ctypes.c_int) -> ctypes.c_bool:
    return _lib.bark_generate_audio(bctx, text, n_threads)

_lib.bark_generate_audio.argtypes = [bark_context_p, ctypes.c_char_p, ctypes.c_int]
_lib.bark_generate_audio.restype = ctypes.c_bool


def bark_get_audio_data(bctx: bark_context_p) -> Pointer[ctypes.c_float]:
    return _lib.bark_get_audio_data(bctx)

_lib.bark_get_audio_data.argtypes = [bark_context_p]
_lib.bark_get_audio_data.restype = ctypes.POINTER(ctypes.c_float)


def bark_get_audio_data_size(bctx: bark_context_p) -> ctypes.c_int:
    return _lib.bark_get_audio_data_size(bctx)

_lib.bark_get_audio_data_size.argtypes = [bark_context_p]
_lib.bark_get_audio_data_size.restype = ctypes.c_int


def bark_get_load_time(bctx: bark_context_p) -> ctypes.c_int64:
    return _lib.bark_get_load_time(bctx)

_lib.bark_get_load_time.argtypes = [bark_context_p]
_lib.bark_get_load_time.restype = ctypes.c_int64


def bark_get_eval_time(bctx: bark_context_p) -> ctypes.c_int64:
    return _lib.bark_get_eval_time(bctx)

_lib.bark_get_eval_time.argtypes = [bark_context_p]
_lib.bark_get_eval_time.restype = ctypes.c_int64


def bark_reset_statistics(bctx: bark_context_p):
    _lib.bark_reset_statistics(bctx)

_lib.bark_reset_statistics.argtypes = [bark_context_p]
_lib.bark_reset_statistics.restype = None


def bark_model_quantize(fname_inp: ctypes.c_char_p,
    fname_out: ctypes.c_char_p,
    ftype: ctypes.c_int) -> ctypes.c_bool:
    return _lib.bark_model_quantize(fname_inp, fname_out, ftype)

_lib.bark_model_quantize.argtypes = [ctypes.c_char_p, ctypes.c_char_p, ctypes.c_int]
_lib.bark_model_quantize.restype = ctypes.c_bool


def bark_free(bctx: bark_context_p):
    _lib.bark_free(bctx)

_lib.bark_free.argtypes = [bark_context_p]
_lib.bark_free.restype = None