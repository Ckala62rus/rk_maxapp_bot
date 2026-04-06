FROM node:20.18-alpine AS builder

WORKDIR /app

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend ./
RUN npm run build

FROM nginx:1.27-alpine

# Копируем статику
COPY --from=builder /app/dist /usr/share/nginx/html

# Копируем готовый конфиг
COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf

CMD ["nginx", "-g", "daemon off;"]