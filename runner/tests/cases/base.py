from typing import Any


class BaseCase:

    def name(self) -> str:
        return self.__class__.__name__

    def param(self) -> list[str]:
        raise NotImplementedError

    def check(self, res: Any) -> bool:  # pyright: ignore[reportUnusedParameter]
        '''
        Optional check function to verify the result.
        '''
        return True
