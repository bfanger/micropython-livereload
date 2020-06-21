TESTS=./pkg/multiplex

ifdef VERBOSE
TEST_PARAM=-v
endif

build:
	go build ./cmd/micropython-livereload

dev:
	find . -name '*.go'|entr -c -r go run ./cmd/micropython-livereload

test:
	go test ${TESTS} ${TEST_PARAM}

test-watch:
	find . -name '*.go'|entr -c -r go test -failfast  ${TESTS} ${TEST_PARAM}


py:
	find .|egrep '(go|py)'|entr -c -r go run ./cmd/micropython-livereload