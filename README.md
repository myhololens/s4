[![Build Status](https://travis-ci.com/axetroy/s4.svg?branch=master)](https://travis-ci.com/axetroy/s4)
[![Coverage Status](https://coveralls.io/repos/github/axetroy/s4/badge.svg?branch=master)](https://coveralls.io/github/axetroy/s4?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/axetroy/s4)](https://goreportcard.com/report/github.com/axetroy/s4)
![License](https://img.shields.io/github/license/axetroy/s4.svg)
![Repo Size](https://img.shields.io/github/repo-size/axetroy/s4.svg)

## s4

Perform remote server tasks on local computer

Features:

- [x] Cross platform support
- [x] Declarative workflow
- [x] Upload local files to remote
- [x] Download remote files to local
- [x] Execute commands on the remote server

### Usage

step 1: create a file name `.s4`

```s4
CONNECT root@192.168.0.1:22

RUN ls -lh
```

step 2: run the following command

```bash
> s4
[Step 1]: CONNECT root@192.168.0.1:22
? Please type remote server's password **********
[Step 2]: RUN ls -lh
total 20K
drwxr-xr-x  4 root root 4.0K Mar 15 10:10 test1
drwxr-xr-x  2 root root 4.0K Sep 23  2018 test2
drwxr-xr-x  6 root root 4.0K Sep 23  2018 test3
drwxr-xr-x  4 root root 4.0K Aug 27 16:25 test4
```

for more detail about command. print `s4 --help`

### Documentation

| Syntax   | Description                                              | Multiple | Example                               |
| -------- | -------------------------------------------------------- | -------- | ------------------------------------- |
| CONNECT  | connect to remote SSH server                             | ✖️       | CONNECT root@192.168.0.1:22           |
| ENV      | set environmental variable for `RUN` command             | ☑️       | ENV PRIVATE_KEY = 123                 |
| CD       | change current working directory of remote server        | ☑️       | CD /home/axetroy                      |
| UPLOAD   | upload local files to remote server                      | ☑️       | UPLOAD start.py ./server              |
| DOWNLOAD | download remote files to local                           | ☑️       | DOWNLOAD start.py ./server            |
| COPY     | copy file on remote server                               | ☑️       | COPY data.db data.db.bak              |
| MOVE     | move file on remote server                               | ☑️       | MOVE data.bak data.db                 |
| DELETE   | delete files on remote server, directory will be ignored | ☑️       | DELETE file1 file2                    |
| RUN      | run command in remote server                             | ☑️       | RUN python ./remote/start.py          |
| CMD      | run command in local server                              | ☑️       | CMD ["cat", "README.md"]              |
| BASH     | run bash script in local server                          | ☑️       | BASH cat package.json \| grep version |

### Installation

Download the executable file for your platform at [release page](https://github.com/axetroy/s4/releases)

Then set the environment variable.

eg, the executable file is in the `~/bin` directory.

```bash
# ~/.bash_profile
export PATH="$PATH:~/bin"
```

finally, try it out.

```bash
s4 --help
```

### Upgrade

You can re-download the executable and overwrite the original file.

or type the following command to upgrade to the latest version.

```bash
> s4 upgrade
```

### Build from source code

```bash
> go get -v -u github.com/axetroy/s4
> cd $GOPATH/src/github.com/axetroy/s4
> make build
> ls -lh ./bin

total 24624
-rw-r--r--  1 axetroy  staff   3.9M Aug 28 14:11 s4_linux_x64.tar.gz
-rw-r--r--@ 1 axetroy  staff   3.8M Aug 28 14:11 s4_osx_x64.tar.gz
-rw-r--r--  1 axetroy  staff   3.8M Aug 28 14:11 s4_win_x64.tar.gz
```

### Test

```bash
make test
```

### Why?

> Why do I need such a tool?
> What is its use?

In development, we need to operate remote servers locally, such as deploying services, restarting services, upload files, etc.

We can of course do this with a bash script.

But that is quite cumbersome.

So, I wrote this tool to release my hands.

I hope this helps you.

### License

The [MIT License](https://github.com/axetroy/s4/blob/master/LICENSE)
