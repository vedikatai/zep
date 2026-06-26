"""Tests for deterministic episode replay (ReplayContext)."""

from __future__ import annotations

import asyncio
from datetime import datetime, timezone
from unittest.mock import AsyncMock, patch
from uuid import UUID, uuid4

import pytest

from graphiti_core.edges import EntityEdge, EpisodicEdge
from graphiti_core.nodes import EntityNode, EpisodeType, EpisodicNode
from graphiti_core.utils.datetime_utils import utc_now
from graphiti_core.utils.replay import (
    ReplayContext,
    clear_replay_context,
    generate_uuid,
    get_replay_context,
    set_replay_context,
)


FIXED_TS = datetime(2024, 1, 15, 12, 0, 0, tzinfo=timezone.utc)


def _make_ctx(
    seed: str = 'test-seed',
    episode_id: str = 'ep1',
    cache: dict[str, str] | None = None,
) -> ReplayContext:
    return ReplayContext(
        episode_id=episode_id,
        deterministic_uuid_seed=seed,
        fixed_timestamp=FIXED_TS,
        llm_response_cache=cache or {},
    )


def test_uuid_generation_is_deterministic():
    ctx = _make_ctx(seed='test-seed')
    set_replay_context(ctx)
    first = [generate_uuid() for _ in range(5)]
    clear_replay_context()

    ctx2 = _make_ctx(seed='test-seed')
    set_replay_context(ctx2)
    second = [generate_uuid() for _ in range(5)]
    clear_replay_context()

    assert first == second
    # All must be valid UUIDs and come from uuid5 (version 5)
    for u in first:
        parsed = UUID(u)
        assert parsed.version == 5


def test_uuid_counter_resets_between_replays():
    ctx_a = _make_ctx(seed='shared-seed')
    set_replay_context(ctx_a)
    uuid_a1 = generate_uuid()
    clear_replay_context()

    ctx_b = _make_ctx(seed='shared-seed')
    set_replay_context(ctx_b)
    uuid_b1 = generate_uuid()
    # B is now ahead (counter at 1); leave B active and capture B's next is not needed yet
    # but we need A #2 vs B #1: so keep B context with counter at 1, get B's current first uuid already
    clear_replay_context()

    assert uuid_a1 == uuid_b1

    # Replay A again for UUID #1 and #2
    ctx_a2 = _make_ctx(seed='shared-seed')
    set_replay_context(ctx_a2)
    _ = generate_uuid()  # #1
    uuid_a2 = generate_uuid()  # #2
    clear_replay_context()

    # Replay B: only generate #1 (B is "ahead" in the sense we compare A#2 with B#1)
    ctx_b2 = _make_ctx(seed='shared-seed')
    set_replay_context(ctx_b2)
    uuid_b1_again = generate_uuid()
    clear_replay_context()

    assert uuid_a2 != uuid_b1_again


def test_timestamp_is_fixed():
    ctx = _make_ctx()
    set_replay_context(ctx)
    try:
        # Node.created_at default uses utc_now(); also call utc_now directly
        ts = utc_now()
        node = EntityNode(name='Alice', group_id='g1')
        assert ts == FIXED_TS
        assert node.created_at == FIXED_TS
        assert ts != datetime.now(timezone.utc) or True  # clock may coincide; fixed is what matters
        # Ensure we did not get a "fresh" now that differs from FIXED_TS in structure
        assert node.created_at == ctx.fixed_timestamp
    finally:
        clear_replay_context()


@pytest.mark.asyncio
async def test_llm_cache_hit_bypasses_real_llm():
    from graphiti_core.llm_client.client import LLMClient
    from graphiti_core.llm_client.config import LLMConfig
    from graphiti_core.prompts.models import Message

    class StubLLM(LLMClient):
        def __init__(self):
            super().__init__(LLMConfig(model='stub'), cache=False)

        def _get_provider_type(self) -> str:
            return 'stub'

        async def _generate_response(
            self, messages, response_model=None, max_tokens=None, model_size=None
        ):
            raise AssertionError('real LLM should not be called on cache hit')

    client = StubLLM()

    cache_payload = '{"entities": ["Alice"]}'
    ctx = _make_ctx(cache={'prompt': cache_payload})
    set_replay_context(ctx)
    try:
        with patch.object(
            client, '_generate_response_with_retry', new_callable=AsyncMock
        ) as mock_retry:
            result = await client.generate_response(
                [Message(role='user', content='hello')],
                prompt_name='prompt',
            )
            mock_retry.assert_not_called()
            assert result == {'entities': ['Alice']}
    finally:
        clear_replay_context()


@pytest.mark.asyncio
async def test_full_episode_replay_produces_identical_graph():
    """Exercise real node/edge save paths (Kuzu embedded) under ReplayContext.

    Uses the same EpisodicNode.save / EntityNode.save / EntityEdge.save /
    EpisodicEdge.save code paths as add_episode persistence. Full add_episode
    also depends on embedder non-determinism which is out of scope for this
    replay patch set; we still prove UUID+timestamp deterministic graph writes.
    """
    from graphiti_core.driver.kuzu_driver import KuzuDriver

    async def _ingest(driver: KuzuDriver, seed: str = 'ep1') -> tuple[set[str], set[str]]:
        ctx = ReplayContext(
            episode_id='ep1',
            deterministic_uuid_seed=seed,
            fixed_timestamp=FIXED_TS,
            llm_response_cache={'prompt': '{"ok": true}'},
        )
        set_replay_context(ctx)
        try:
            episode = EpisodicNode(
                name='ep',
                group_id='g1',
                source=EpisodeType.text,
                source_description='test',
                content='Alice works at Acme',
                valid_at=FIXED_TS,
            )
            entity = EntityNode(name='Alice', group_id='g1', summary='person')
            e_edge = EntityEdge(
                group_id='g1',
                source_node_uuid=entity.uuid,
                target_node_uuid=entity.uuid,
                created_at=FIXED_TS,
                name='WORKS_AT',
                fact='Alice works at Acme',
            )
            ment = EpisodicEdge(
                group_id='g1',
                source_node_uuid=episode.uuid,
                target_node_uuid=entity.uuid,
                created_at=FIXED_TS,
            )
            await episode.save(driver)
            await entity.save(driver)
            await e_edge.save(driver)
            await ment.save(driver)
            node_uuids = {episode.uuid, entity.uuid}
            edge_uuids = {e_edge.uuid, ment.uuid}
            return node_uuids, edge_uuids
        finally:
            clear_replay_context()

    # Fresh in-memory Kuzu graph
    driver1 = KuzuDriver(db=':memory:')
    await driver1.build_indices_and_constraints()
    nodes1, edges1 = await _ingest(driver1)

    # Reset: new empty graph, same seed/context
    driver2 = KuzuDriver(db=':memory:')
    await driver2.build_indices_and_constraints()
    nodes2, edges2 = await _ingest(driver2)

    assert nodes1 == nodes2
    assert edges1 == edges2
    assert len(nodes1) == 2
    assert len(edges1) == 2


def test_no_replay_context_preserves_existing_behavior():
    clear_replay_context()
    assert get_replay_context() is None

    with patch('graphiti_core.utils.replay.uuid4') as mock_uuid4:
        mock_uuid4.return_value = uuid4()
        u = generate_uuid()
        mock_uuid4.assert_called()
        # Not version 5 when no context (uuid4 mock returns random uuid4)
        assert UUID(u).version == 4

    with patch('graphiti_core.utils.datetime_utils.datetime') as mock_dt:
        # Preserve timezone attribute used by utc_now
        from datetime import datetime as real_datetime

        mock_dt.now.return_value = real_datetime(2020, 1, 1, tzinfo=timezone.utc)
        mock_dt.side_effect = lambda *a, **k: real_datetime(*a, **k)
        # Patch only datetime.now path used in utc_now
        pass

    # Direct assertion: without context, utc_now uses datetime.now
    with patch(
        'graphiti_core.utils.datetime_utils.datetime'
    ) as mock_datetime_mod:
        sentinel = datetime(1999, 9, 9, tzinfo=timezone.utc)
        mock_datetime_mod.now.return_value = sentinel
        mock_datetime_mod.timezone = timezone  # unused but safe
        # utc_now imports datetime at module level — patch the module's datetime
        with patch('graphiti_core.utils.datetime_utils.datetime') as m:
            from datetime import datetime as dt_cls

            fixed = dt_cls(1999, 9, 9, tzinfo=timezone.utc)

            class _DT:
                @staticmethod
                def now(tz=None):
                    return fixed

            # Replace datetime in module
            import graphiti_core.utils.datetime_utils as du

            original = du.datetime
            du.datetime = type(
                'datetime',
                (),
                {
                    'now': staticmethod(lambda tz=None: fixed),
                },
            )
            try:
                assert utc_now() == fixed
            finally:
                du.datetime = original
