FROM node:20-alpine AS builder

ARG VITE_BASE_PATH=/
ARG VITE_ANALYTICS_SCRIPT_URL
ARG VITE_ANALYTICS_SITE_ID
ENV VITE_BASE_PATH=$VITE_BASE_PATH
ENV VITE_ANALYTICS_SCRIPT_URL=$VITE_ANALYTICS_SCRIPT_URL
ENV VITE_ANALYTICS_SITE_ID=$VITE_ANALYTICS_SITE_ID

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
