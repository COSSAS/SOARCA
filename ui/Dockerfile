FROM node:20.11.1-alpine3.19 AS build

RUN addgroup -g 1001 appgroup && adduser -u 1001 -G appgroup -D appuser

USER appuser

WORKDIR /app

COPY --chown=appuser:appgroup package.json ./

RUN yarn install --frozen-lockfile --production

ENV PATH=/app/node_modules/.bin:$PATH

COPY --chown=appuser:appgroup . .

RUN yarn run build

FROM nginx:1.27-alpine AS prod

COPY --chown=nginx:nginx ./nginx.conf /etc/nginx/conf.d/default.conf

WORKDIR /var/www/html

RUN rm -rf ./*

COPY --from=build /app/dist /var/www/html/
RUN chmod -R 755 /var/www/html && chown -R nginx:nginx /var/www/html


EXPOSE 5173


ENTRYPOINT ["nginx", "-g", "daemon off;"]
