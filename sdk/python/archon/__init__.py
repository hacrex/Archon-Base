"""Archon Base Python SDK (skeleton).

The real SDK will resolve bindings (vector, storage, models) injected by
the platform at runtime. These stubs define the intended public shape.
"""

from typing import Any, Awaitable, Callable


class _Agent:
    def handler(self, fn: Callable[..., Awaitable[Any]]):
        fn.__archon_handler__ = True
        return fn


agent = _Agent()


class _VectorClient:
    def __init__(self, instance: str):
        self.instance = instance

    async def query(self, text: str, top_k: int = 5):
        raise NotImplementedError("vector bindings are not implemented yet")


class _LLMClient:
    def __init__(self, alias: str):
        self.alias = alias

    async def chat(self, messages, stream: bool = False):
        raise NotImplementedError("model gateway is not implemented yet")


def vector(instance: str) -> _VectorClient:
    return _VectorClient(instance)


def llm(alias: str) -> _LLMClient:
    return _LLMClient(alias)
