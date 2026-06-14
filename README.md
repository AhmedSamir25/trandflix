# trandflix

## AI provider

The AI assistant and chat support both OpenRouter and OpenAI-compatible APIs.

Use OpenRouter:

```env
AI_PROVIDER=openrouter
OPENROUTER_API_KEY=your_openrouter_key
OPENROUTER_MODEL=openai/gpt-4o-mini
```

Use an OpenAI-compatible provider such as DeepSeek or GLM:

```env
AI_PROVIDER=openai_compatible
OPENAI_COMPATIBLE_API_KEY=your_provider_key
OPENAI_COMPATIBLE_BASE_URL=https://api.deepseek.com/v1
OPENAI_COMPATIBLE_MODEL=deepseek-chat
```

`OPENAI_COMPATIBLE_BASE_URL` can be either the base API URL, such as `https://api.deepseek.com/v1`, or the full chat completions URL.
