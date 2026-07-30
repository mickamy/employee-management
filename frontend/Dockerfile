# remix/node-tsx runs TypeScript directly, so dev dependencies stay in the
# runtime image and there is no build step.
FROM node:26-slim
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
ENV NODE_ENV=production
EXPOSE 44100
CMD ["npm", "run", "start"]
