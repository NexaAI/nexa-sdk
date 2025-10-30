from typing import final, override

from .base import BaseCase


@final
class QueryDocument(BaseCase):

    @override
    def param(self) -> list[str]:
        return [
            '-q',
            'What is the capital of France?',
            '-d',
            'Paris is the capital of France. It is known for its art, culture, and history.',
            '-d',
            'London is the capital of the United Kingdom. It is famous for its landmarks and museums.',
            '-d',
            'Berlin is the capital of Germany. It has a rich history and vibrant culture.',
        ]
