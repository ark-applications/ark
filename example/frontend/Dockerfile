FROM node:20-alpine

WORKDIR /app/

COPY package.json /app
COPY yarn.lock /app

RUN yarn

COPY public/ /app/public
COPY src/ /app/src

CMD ["yarn", "start"]
