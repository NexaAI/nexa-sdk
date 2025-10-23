from typing import final, override

from .base import BaseCase


@final
class SingleRound(BaseCase):

    @override
    def param(self) -> list[str]:
        return [
            '-p',
            'hi, how are you today?',
        ]
