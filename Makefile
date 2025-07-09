run:
	@clear
	cd backend && go run .
.PHONY: frontend
frontend:
	cd frontend && npm install &&  npm run dev
messenger:
	cd desktop-messenger && npm install && npm start
format:
	cd backend && gofmt -w -s .
	@echo "files are formatted correctly"

build-frontend:
	cd frontend && npm install && npm run build
	@echo "frontend is built successfully. Running on port 3000"