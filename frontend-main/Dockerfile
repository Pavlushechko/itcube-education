FROM node:20-alpine

WORKDIR /app

# зависимости отдельно (кешируется)
COPY package.json package-lock.json* ./
RUN npm ci

# код
COPY . .

EXPOSE 3000
CMD ["npm", "run", "dev", "--", "-H", "0.0.0.0", "-p", "3000"]
