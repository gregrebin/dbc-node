
# DATA BLOCKCHAIN

Official implementation of Data Blockchain (DBC).

This is a Proof of Concept, for internal usage only. Functionality may change.
The app is unstable.

### Project objectives
* Give a secure way to transmit data privately on a public ledger, from provider
to requirer.
* Give a way to proof that the part providing data is a real cerified entity
(ex. has an official id) to everyone without revealing any private information about
such entity.
* Deterministically verify that the provided data satisfies the requested parameter 
without revealing any data content to anyone, besides the requirer and the
provider (data examples: survey, news, private message, medical certificate etc.).

### Implementation
...

### Installation
DBC is written is Go and for now can be only compiled from source:

Install the Go tool as specified in [Getting Started](https://golang.org/doc/install).

Clone this directory using git to directory of your choosing:

```shell
git clone ...
```

Use install to compile DBC binary to GOBIN directory

```shell
cd dbc-node
go install
```

Or use build to compile it to the current directory

```shell
cd dbc-node
go build
```

For more information check [How to Write Go Code](https://golang.org/doc/code.html).

### Usage
Before running DBC the first time you should create a home directory with
 
```shell script
dbc-node init [--home]
```

After you can run the node using

```shell script
dbc-node run [--home]
```

The application stores configuration and data inside home directory that can be
specified with `--home` flag for both `init` and `run` commands.