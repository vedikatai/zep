"""Deterministic episode replay support for graphiti_core.

When a ReplayContext is active (via set_replay_context / context manager),
UUID generation, timestamps, and LLM responses become deterministic so that
replaying the same episode yields identical node and edge UUIDs.
"""

from __future__ import annotations

import contextvars
from contextlib import contextmanager
from dataclasses import dataclass, field
from datetime import datetime
from typing import Iterator
from uuid import NAMESPACE_DNS, uuid4, uuid5

_replay_context: contextvars.ContextVar[ReplayContext | None] = contextvars.ContextVar(
    'graphiti_replay_context', default=None
)


@dataclass
class ReplayContext:
    """Active deterministic replay configuration.

    Attributes:
        episode_id: Identifier for the episode being replayed.
        deterministic_uuid_seed: Seed mixed into uuid5 name generation.
        fixed_timestamp: Timestamp returned by utc_now() while active.
        llm_response_cache: Map of cache key (prompt_name or prompt text) to
            JSON-serializable response string (or serialized dict).
        _uuid_counter: Internal counter; increments per UUID generated.
            Reset to 0 when the context is set active.
    """

    episode_id: str
    deterministic_uuid_seed: str
    fixed_timestamp: datetime
    llm_response_cache: dict[str, str] = field(default_factory=dict)
    _uuid_counter: int = field(default=0, init=False, repr=False)

    def next_uuid(self) -> str:
        """Return the next deterministic UUID and advance the counter."""
        name = f'{self.deterministic_uuid_seed}:{self._uuid_counter}'
        self._uuid_counter += 1
        return str(uuid5(NAMESPACE_DNS, name))

    def reset_counter(self) -> None:
        """Reset the UUID counter to 0 (start of a replay)."""
        self._uuid_counter = 0


def set_replay_context(ctx: ReplayContext | None) -> contextvars.Token:
    """Set the active ReplayContext (thread- and async-safe via ContextVar).

    Resets the UUID counter to 0 when a non-None context is set so each
    replay starts from a clean counter sequence.
    """
    if ctx is not None:
        ctx.reset_counter()
    return _replay_context.set(ctx)


def get_replay_context() -> ReplayContext | None:
    """Return the active ReplayContext, or None if not replaying."""
    return _replay_context.get()


def clear_replay_context() -> None:
    """Clear any active ReplayContext."""
    _replay_context.set(None)


@contextmanager
def replay_context(ctx: ReplayContext) -> Iterator[ReplayContext]:
    """Context manager that activates a ReplayContext for the block."""
    token = set_replay_context(ctx)
    try:
        yield ctx
    finally:
        _replay_context.reset(token)


def generate_uuid() -> str:
    """Generate a UUID: deterministic uuid5 under replay, else uuid4."""
    ctx = get_replay_context()
    if ctx is not None:
        return ctx.next_uuid()
    return str(uuid4())


def lookup_llm_cache(key: str) -> str | None:
    """Return a cached LLM response string if present in the active context."""
    ctx = get_replay_context()
    if ctx is None:
        return None
    return ctx.llm_response_cache.get(key)


def check_replay_llm_cache(
    messages: list,
    prompt_name: str | None = None,
) -> dict | None:
    """If ReplayContext is active and a cache entry matches, return a parsed dict.

    Cache keys tried (in order): prompt_name (if set), then joined message contents.
    Values may be JSON objects (returned as dict) or plain strings (wrapped as
    ``{'content': ...}``).
    """
    import json

    ctx = get_replay_context()
    if ctx is None:
        return None

    cache_keys: list[str] = []
    if prompt_name:
        cache_keys.append(prompt_name)
    try:
        joined = '\n'.join(m.content for m in messages)
        cache_keys.append(joined)
    except Exception:
        pass

    for key in cache_keys:
        cached = ctx.llm_response_cache.get(key)
        if cached is None:
            continue
        try:
            parsed = json.loads(cached)
            if isinstance(parsed, dict):
                return parsed
        except (json.JSONDecodeError, TypeError):
            pass
        return {'content': cached}
    return None
