# @testra/sdk

Official Testra TypeScript SDK.

Types are generated from `docs/api/openapi/openapi.yaml` with `openapi-typescript`.
Run `pnpm --filter @testra/sdk generate` to regenerate them.

```ts
import { createTestraClient } from "@testra/sdk";

const client = createTestraClient("https://api.testra.io/api/v1");
const { data, error } = await client.GET("/auth/me");
```
