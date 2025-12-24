from typing import final, override

from .base import BaseCase


@final
class ImageRecognition(BaseCase):

    @override
    def param(self) -> list[str]:
        return ['-i', './assets/cat.jpg']


@final
class OCR(BaseCase):

    @override
    def param(self) -> list[str]:
        return ['-i', './assets/text.jpeg']
