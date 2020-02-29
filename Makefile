
build:
	go build ./cmd/micropython-livereload

dev:
	find .|egrep '(go|py)'|entr -c -r go run ./cmd/micropython-livereload