TESTS=./pkg/multiplex
TEST_PARAM=-run "^TestRead"
ifdef VERBOSE
TEST_PARAM=-v
endif

build:
	go build ./cmd/micropython-livereload

dev:
	find . -name '*.go'|entr -c -r go run ./cmd/micropython-livereload

test:
	go test ${TESTS} ${TEST_PARAM}
	python py/multiplex.test.py

test-watch:
	find . -name '*.go'|entr -c -r go test -failfast  ${TESTS} ${TEST_PARAM}

py:
	find .|egrep '(go|py)'|entr -c -r go run ./cmd/micropython-livereload
py-test-watch:
	find . -name '*.py'|entr -c -r python py/multiplex.test.py