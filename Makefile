.PHONY: start test lint

start:
	node server.js

test:
	mocha test/index.js

lint:
	./node_modules/.bin/eslint server.js test/ src/ --fix
