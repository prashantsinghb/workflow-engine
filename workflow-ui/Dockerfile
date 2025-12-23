# Stage 0, "build-stage", based on Node.js, to build and compile the frontend
FROM node:20.17-alpine AS build-stage

WORKDIR /app

# inatall packages & cache
COPY package*.json /app/
RUN npm i

# copy code and build
COPY ./ /app/
RUN npm run build


# Stage 1, based on Nginx, to have only the compiled app, ready for production with Nginx
FROM nginxinc/nginx-unprivileged:1.26-alpine-slim
ARG BUILD_TAG=dev

ENV VITE_API_BASE_URL=http://192.168.100.103:31200

COPY --from=build-stage /app/dist/ /usr/share/nginx/html

COPY ./nginx.conf /etc/nginx/conf.d/default.conf
COPY ./start-nginx.sh /usr/bin/start-nginx.sh
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
COPY ./start-nginx.sh /usr/bin/start-nginx.sh

USER root
RUN chmod -R 777 /usr/share/nginx/html/ && chmod -R 777 /tmp && chmod -R 777 /var
USER nginx

# append TCP forwarding constructs to nginx.conf
RUN echo "stream {" >> /etc/nginx/nginx.conf && \
    echo "    include /etc/nginx/tcp_conf.d/*.conf;" >> /etc/nginx/nginx.conf && \
    echo "}" >> /etc/nginx/nginx.conf

ENTRYPOINT [ "start-nginx.sh" ]
