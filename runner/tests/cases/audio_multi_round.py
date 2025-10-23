from typing import final, override

from .base import BaseCase


@final
class AudioMultiRound(BaseCase):

    @override
    def param(self) -> list[str]:
        return [
            '-p',
            'transcribe the audio ./assets/osr_us.wav',
            '-p',
            'transcribe the audio ./assets/osr_us.wav',
        ]
