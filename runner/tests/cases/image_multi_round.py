from typing import final, override

from .base import BaseCase


@final
class ImageMultiRound(BaseCase):

    @override
    def param(self) -> list[str]:
        return [
            '-p',
            'describe the image ./assets/text.png',
            '-p',
            'compare the two images ./assets/text.png and ./assets/cat.jpg',
        ]
