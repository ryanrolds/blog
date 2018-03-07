.PHONY: start test lint

start:
	node server.js

test:
	mocha test/index.js

lint:
	./node_modules/.bin/eslint server.js test/ src/ --fix

migrate-create:
	./node_modules/.bin/db-migrate create name=$(NAME)

migrate-up:
	./node_modules/.bin/db-migrate up

migrate-down:
	./node_modules/.bin/db-migrate down
