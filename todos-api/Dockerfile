FROM node:12-alpine

EXPOSE 8082

WORKDIR /usr/src/app

COPY package.json ./
RUN npm install
RUN npm install newrelic --save

COPY . .

CMD ["sh", "-c", "npm start" ]
