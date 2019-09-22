## Setup

micropython with VfsFat enabled

### Docker

```
docker build --tag micropython-coverage .
alias mpy="docker run -v $PWD:/app --rm -it micropython-coverage"
```

### macOS

Download full source from: http://www.micropython.org/download

```
brew install libffi
export PKG_CONFIG_PATH="/usr/local/opt/libffi/lib/pkgconfig"

cd micropython-?.??/ports/unix
make coverage
```
