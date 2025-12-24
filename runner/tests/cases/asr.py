from typing import final, override

from .base import BaseCase


@final
class ASR(BaseCase):

    @override
    def param(self) -> list[str]:
        return ['-i', './assets/storytelling.wav']
