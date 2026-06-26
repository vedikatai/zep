# Python integration test coverage matrix

Public = module-level or class method not starting with `_` under `integrations/*/python/src/`.
`has_direct_test` = some `tests/**/test_*.py` contains an AST `Call` to that function/method name.

| integration | function | has_direct_test | test file |
|---|---|---|---|
| adk | `ZepContextTool.process_llm_request` | yes | integrations/adk/python/tests/test_basic.py |
| adk | `ZepGraphSearchTool.run_async` | yes | integrations/adk/python/tests/test_basic.py |
| adk | `create_after_model_callback` | yes | integrations/adk/python/tests/test_basic.py |
| ag2 | `ZepMemoryManager.client` | no |  |
| ag2 | `ZepMemoryManager.user_id` | no |  |
| ag2 | `ZepMemoryManager.session_id` | no |  |
| ag2 | `ZepMemoryManager.get_memory_context` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `ZepMemoryManager.enrich_system_message` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `ZepMemoryManager.add_messages` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `ZepMemoryManager.get_session_facts` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `ZepMemoryManager.get_memory_context_sync` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `ZepMemoryManager.enrich_system_message_sync` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `shutdown_background_loop` | no |  |
| ag2 | `create_search_memory_tool` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `create_add_memory_tool` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `create_search_graph_tool` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `create_add_graph_data_tool` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `register_all_tools` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `ZepGraphMemoryManager.client` | no |  |
| ag2 | `ZepGraphMemoryManager.graph_id` | no |  |
| ag2 | `ZepGraphMemoryManager.search` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `ZepGraphMemoryManager.add_data` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `ZepGraphMemoryManager.enrich_system_message` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `ZepGraphMemoryManager.search_sync` | yes | integrations/ag2/python/tests/test_basic.py |
| ag2 | `ZepGraphMemoryManager.add_data_sync` | yes | integrations/ag2/python/tests/test_basic.py |
| autogen | `ZepUserMemory.add` | yes | integrations/autogen/python/tests/test_basic.py |
| autogen | `ZepUserMemory.query` | yes | integrations/autogen/python/tests/test_basic.py |
| autogen | `ZepUserMemory.update_context` | yes | integrations/autogen/python/tests/test_basic.py |
| autogen | `ZepUserMemory.clear` | no |  |
| autogen | `ZepUserMemory.close` | no |  |
| autogen | `search_memory` | no |  |
| autogen | `add_graph_data` | no |  |
| autogen | `create_search_graph_tool` | no |  |
| autogen | `create_add_graph_data_tool` | no |  |
| autogen | `ZepGraphMemory.add` | yes | integrations/autogen/python/tests/test_basic.py |
| autogen | `ZepGraphMemory.query` | yes | integrations/autogen/python/tests/test_basic.py |
| autogen | `ZepGraphMemory.update_context` | yes | integrations/autogen/python/tests/test_basic.py |
| autogen | `ZepGraphMemory.clear` | no |  |
| autogen | `ZepGraphMemory.close` | no |  |
| crewai | `ZepUserStorage.save` | yes | integrations/crewai/python/tests/test_basic.py |
| crewai | `ZepUserStorage.search` | yes | integrations/crewai/python/tests/test_basic.py |
| crewai | `ZepUserStorage.get_context` | yes | integrations/crewai/python/tests/test_user_storage.py |
| crewai | `ZepUserStorage.reset` | yes | integrations/crewai/python/tests/test_basic.py |
| crewai | `ZepUserStorage.user_id` | no |  |
| crewai | `ZepUserStorage.thread_id` | no |  |
| crewai | `ZepStorage.save` | yes | integrations/crewai/python/tests/test_basic.py |
| crewai | `ZepStorage.search` | yes | integrations/crewai/python/tests/test_basic.py |
| crewai | `ZepStorage.reset` | yes | integrations/crewai/python/tests/test_basic.py |
| crewai | `ZepStorage.user_id` | no |  |
| crewai | `ZepStorage.thread_id` | no |  |
| crewai | `ZepSearchTool.client` | no |  |
| crewai | `ZepSearchTool.graph_id` | no |  |
| crewai | `ZepSearchTool.user_id` | no |  |
| crewai | `ZepAddDataTool.client` | no |  |
| crewai | `ZepAddDataTool.graph_id` | no |  |
| crewai | `ZepAddDataTool.user_id` | no |  |
| crewai | `create_search_tool` | yes | integrations/crewai/python/tests/test_integration.py |
| crewai | `create_add_data_tool` | yes | integrations/crewai/python/tests/test_tools.py |
| crewai | `search_graph_and_compose_context` | no |  |
| crewai | `ZepGraphStorage.save` | yes | integrations/crewai/python/tests/test_basic.py |
| crewai | `ZepGraphStorage.search` | yes | integrations/crewai/python/tests/test_basic.py |
| crewai | `ZepGraphStorage.reset` | yes | integrations/crewai/python/tests/test_basic.py |
| crewai | `ZepGraphStorage.graph_id` | no |  |
| langgraph | `ZepStore.batch` | yes | integrations/langgraph/python/tests/test_store.py |
| langgraph | `ZepStore.abatch` | no |  |
| langgraph | `to_zep_message` | yes | integrations/langgraph/python/tests/test_persistence.py |
| langgraph | `to_zep_messages` | yes | integrations/langgraph/python/tests/test_persistence.py |
| langgraph | `persist_messages` | yes | integrations/langgraph/python/tests/test_persistence.py |
| langgraph | `persist_messages_sync` | yes | integrations/langgraph/python/tests/test_persistence.py |
| langgraph | `create_graph_search_tool` | yes | integrations/langgraph/python/tests/test_integration.py |
| langgraph | `create_graph_search_tool_sync` | yes | integrations/langgraph/python/tests/test_tools.py |
| langgraph | `get_zep_context` | yes | integrations/langgraph/python/tests/test_context.py |
| langgraph | `get_zep_context_sync` | yes | integrations/langgraph/python/tests/test_context.py |
| langgraph | `format_context_block` | yes | integrations/langgraph/python/tests/test_context.py |
| langgraph | `build_system_message` | yes | integrations/langgraph/python/tests/test_context.py |
| langgraph | `build_system_message_sync` | yes | integrations/langgraph/python/tests/test_context.py |
| livekit | `ZepUserAgent.on_enter` | no |  |
| livekit | `ZepUserAgent.on_user_turn_completed` | yes | integrations/livekit/python/tests/test_integration.py |
| livekit | `ZepUserAgent.on_exit` | no |  |
| livekit | `ZepGraphAgent.on_enter` | no |  |
| livekit | `ZepGraphAgent.on_user_turn_completed` | yes | integrations/livekit/python/tests/test_integration.py |
| livekit | `ZepGraphAgent.on_exit` | no |  |
| ms-agent-framework | `truncate_message_content` | yes | integrations/ms-agent-framework/python/tests/test_basic.py |
| ms-agent-framework | `ZepContextProvider.user_id` | no |  |
| ms-agent-framework | `ZepContextProvider.thread_id` | no |  |
| ms-agent-framework | `ZepContextProvider.before_run` | yes | integrations/ms-agent-framework/python/tests/test_basic.py |
| ms-agent-framework | `ZepContextProvider.after_run` | yes | integrations/ms-agent-framework/python/tests/test_basic.py |
| pydantic-ai | `ZepDeps.display_name` | no |  |
| pydantic-ai | `truncate_message_content` | no |  |
| pydantic-ai | `latest_user_text` | yes | integrations/pydantic-ai/python/tests/test_deps.py |
| pydantic-ai | `model_messages_to_zep` | yes | integrations/pydantic-ai/python/tests/test_deps.py |
| pydantic-ai | `make_context_request` | yes | integrations/pydantic-ai/python/tests/test_deps.py |
| pydantic-ai | `ensure_user_and_thread` | yes | integrations/pydantic-ai/python/tests/test_deps.py |
| pydantic-ai | `create_zep_search_tool` | yes | integrations/pydantic-ai/python/tests/test_search.py |
| pydantic-ai | `zep_history_processor` | yes | integrations/pydantic-ai/python/tests/test_history_processor.py |
| pydantic-ai | `persist_run` | yes | integrations/pydantic-ai/python/tests/test_history_processor.py |
| pydantic-ai | `reset_turn_cache` | yes | integrations/pydantic-ai/python/tests/test_history_processor.py |
