FROM node:20.18-alpine AS builder

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend ./
RUN npm run build

FROM nginx:1.27-alpine

# Default (HTTP-only) nginx config; entrypoint switches to SSL config
# when certificate files exist inside the container.
COPY deploy/nginx/nginx.conf /etc/nginx/conf.d/default.conf
COPY deploy/nginx/nginx.ssl.template.conf /etc/nginx/templates/nginx.ssl.template.conf
COPY deploy/nginx/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh
COPY --from=builder /app/dist /usr/share/nginx/html

ENTRYPOINT ["/entrypoint.sh"]
EXPOSE 80
