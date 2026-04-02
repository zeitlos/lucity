FROM node:20-alpine AS builder

ARG VITE_BASE_PATH=/
ARG VERSION=dev
ENV VITE_BASE_PATH=$VITE_BASE_PATH
ENV APP_VERSION=$VERSION

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
