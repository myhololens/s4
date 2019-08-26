## s4

defined the jobs for remote and do it at local.

### Usage

create a file name `.s4`

```s4
HOST 192.168.0.1 # your remote server IP

PORT 2222 # SSH port

USERNAME root # username for the host

CWD /root # set current work dir

COPY ./README.md ./test dir # copy local `./README` to the remote `/root/test`

RUN cat ./test/README.md # run the command
```

```bash
$ s4 --help
```

### Documentation

| Keyword  | Description                            | Multiple selection |
| -------- | -------------------------------------- | ------------------ |
| HOST     | remote ssh server address              | ✖️                 |
| PORT     | remote ssh server port                 | ✖️                 |
| USERNAME | remote ssh server username             | ✖️                 |
| CWD      | set current work dir for remote server | ☑️                 |
| COPY     | copy local files to remote server      | ☑️                 |
| RUN      | run command in remote command          | ☑️                 |

### Build from source code

```bash
go get -v -u github.com/axetroy/s4
cd $GOPATH/src/github.com/axetroy/s4

make build
ls -lh ./bin

```

### Test

```bash
make test
```

### License

The [MIT License](https://github.com/axetroy/s4/blob/master/LICENSE)
