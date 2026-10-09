ARG NEXT_PUBLIC_API_URL=http://localhost:8085
FROM node:20-alpine AS build
WORKDIR /app
ARG NEXT_PUBLIC_API_URL
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
COPY package.json ./
RUN npm install
COPY . .
RUN mkdir -p public
RUN npm run build

FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production
COPY --from=build /app/package.json ./
COPY --from=build /app/.next ./.next
COPY --from=build /app/public ./public
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/next.config.js ./next.config.js
EXPOSE 3000
CMD ["npx", "next", "start", "-p", "3000"]
