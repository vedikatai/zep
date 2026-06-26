# Environment variable audit

## Documented and used (in code and in docs)

- `ANTHROPIC_API_KEY` — code: examples/python/claude-prompt-caching-example/benchmark.py, examples/python/claude-prompt-caching-example/chat.py; docs: examples/python/claude-prompt-caching-example/.env.example
- `GOOGLE_API_KEY` — code: integrations/adk/go/examples/main.go, integrations/adk/python/examples/basic_agent.py, integrations/adk/python/tests/test_integration.py, zep-eval-harness/zep_chunk_documents.py, zep-eval-harness/zep_evaluate.py; docs: integrations/adk/go/SETUP.md, integrations/adk/python/README.md, integrations/adk/python/SETUP.md, integrations/adk/typescript/SETUP.md, zep-eval-harness/.env.example
- `LOG_LEVEL` — code: mcp/zep-mcp-server/internal/config/config_test.go; docs: .env.example, mcp/zep-mcp-server/.env.example, mcp/zep-mcp-server/README.md
- `OPENAI_API_KEY` — code: benchmarks/locomo/benchmark.py, benchmarks/longmemeval/Zep Test Harness/zep_eval.py, benchmarks/longmemeval/Zep Test Harness/zep_responses.py, benchmarks/longmemeval/zep_longmem_eval.py, examples/go/chunking-example/main.go; docs: examples/go/chunking-example/.env.example, examples/python/agent-memory-full-example/.env.example, examples/python/chunking-example/.env.example, examples/python/context-templates-example/.env.example, examples/python/context-templates-example/README.md
- `OPENAI_MODEL` — code: integrations/ag2/python/tests/test_integration.py, integrations/autogen/python/tests/test_integration.py, integrations/crewai/python/tests/test_integration.py, integrations/langgraph/python/tests/test_integration.py, integrations/ms-agent-framework/python/examples/basic_agent.py; docs: integrations/autogen/python/SETUP.md, integrations/crewai/python/SETUP.md, integrations/ms-agent-framework/python/SETUP.md
- `PORT` — code: examples/python/elevenlabs-zep-example/llm-proxy/proxy_server.py; docs: examples/python/elevenlabs-zep-example/llm-proxy/.env.example
- `PROXY_API_KEY` — code: examples/python/elevenlabs-zep-example/llm-proxy/proxy_server.py, examples/python/elevenlabs-zep-example/llm-proxy/test_proxy_locally.py; docs: examples/python/elevenlabs-zep-example/README.md, examples/python/elevenlabs-zep-example/llm-proxy/.env.example
- `ZEP_API_KEY` — code: benchmarks/locomo/benchmark.py, benchmarks/longmemeval/zep_longmem_eval.py, examples/go/chunking-example/main.go, examples/go/entity_types.go, examples/go/user_graph.go; docs: .env.example, benchmarks/locomo/README.md, examples/go/chunking-example/.env.example, examples/python/agent-memory-full-example/.env.example, examples/python/chunking-example/.env.example
- `ZEP_CONFIG_FILE` — code: legacy/src/lib/config/load_ce.go; docs: legacy/.env.example

## Used in code but not documented

(none after 5E documentation updates)

## Documented but not used in code

- `CHUNK_OVERLAP` — zep-eval-harness/README.md
- `CHUNK_SIZE` — zep-eval-harness/README.md
- `DOC_ENTITIES_LIMIT` — zep-eval-harness/README.md
- `DOC_EPISODES_LIMIT` — zep-eval-harness/README.md
- `DOC_FACTS_LIMIT` — zep-eval-harness/README.md
- `LIVEKIT_API_KEY` — integrations/livekit/python/README.md, integrations/livekit/python/SETUP.md, integrations/livekit/python/examples/full-example/.env.example
- `LIVEKIT_API_SECRET` — integrations/livekit/python/README.md, integrations/livekit/python/SETUP.md, integrations/livekit/python/examples/full-example/.env.example
- `LIVEKIT_URL` — integrations/livekit/python/README.md, integrations/livekit/python/SETUP.md, integrations/livekit/python/examples/full-example/.env.example
- `LLM_CONTEXTUALIZATION_MODEL` — zep-eval-harness/README.md
- `LLM_JUDGE_MODEL` — zep-eval-harness/README.md
- `LLM_RESPONSE_MODEL` — zep-eval-harness/README.md
- `STATE_KEYS` — integrations/adk/typescript/README.md
- `TAVILY_API_KEY` — examples/typescript/langgraph/.env.example
- `USER_ENTITIES_LIMIT` — zep-eval-harness/README.md
- `USER_EPISODES_LIMIT` — zep-eval-harness/README.md
- `USER_FACTS_LIMIT` — zep-eval-harness/README.md
- `VITE_ELEVENLABS_AGENT_ID` — examples/python/elevenlabs-zep-example/react-app/.env.example
