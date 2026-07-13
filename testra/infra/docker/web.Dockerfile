FROM node:20-alpine AS builder
RUN corepack enable && corepack prepare pnpm@9.5.0 --activate
WORKDIR /app
COPY . .
RUN pnpm install --frozen-lockfile
RUN pnpm --filter @testra/web build

FROM node:20-alpine
RUN corepack enable && corepack prepare pnpm@9.5.0 --activate
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/apps/web/.next/standalone ./
COPY --from=builder /app/apps/web/.next/static ./apps/web/.next/static
COPY --from=builder /app/apps/web/public ./apps/web/public
EXPOSE 3000
CMD ["node", "apps/web/server.js"]
