FROM node:20-alpine AS build

ARG NUXT_PUBLIC_ANALYTICS_SCRIPT_URL
ARG NUXT_PUBLIC_ANALYTICS_SITE_ID

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run generate

FROM nginx:alpine

COPY --from=build /app/.output/public /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
