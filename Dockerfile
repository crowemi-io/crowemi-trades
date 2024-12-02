FROM node:alpine AS builder

ENV BUILD="true"

WORKDIR /app

COPY package*.json ./

RUN npm install

COPY . .

RUN npm run build

FROM node:alpine AS runner

WORKDIR /app

COPY --from=builder /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/public ./public
COPY --from=builder /app/package.json ./package.json
COPY --from=builder /app/app ./app
COPY --from=builder /app/next.config.js ./next.config.js

ENV NODE_ENV production

EXPOSE 3000

CMD ["npm", "start"]