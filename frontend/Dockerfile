# Filter out irrelevant files for frontend build to improve caching
FROM alpine:latest as builder-frontend-prepare
COPY . /tmp/build
RUN cd /tmp/build && rm -rf caddy Dockerfile


# Build frontend in separate container
FROM node:18-alpine as builder-frontend

RUN apk add --no-cache git git-lfs
RUN mkdir /build
WORKDIR /build

# Download dependencies independently for faster build
COPY package.json package-lock.json ./
RUN npm ci

# Build application
COPY --from=builder-frontend-prepare /tmp/build/ ./
RUN npm run build


# Final container
FROM caddy:2.7.4-alpine

COPY caddy/Caddyfile /etc/caddy/
COPY --from=builder-frontend /build/out /srv

EXPOSE 80 443

ENTRYPOINT ["/usr/bin/caddy"]
CMD ["run", "--config", "/etc/caddy/Caddyfile"]
