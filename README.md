
# DATA BLOCKCHAIN

Data blockchain is a decentralized data exchange platform, that stores encrypted 
user data on a proof-of-stake blockchain. DBC enables users to require and provide 
specific types of data, such as surveys, documents, invoices, etc. over a 
decentralized network, proof that the part providing data is a certified entity
with a zero-knowledge-proof algorithm, automate the verification of such data, 
and the consequent actions, such as payment. DBC node is the reference 
implementation written in Golang, using Tendermint middleware. 

**This is a Proof of Concept, for internal usage only. Functionality may change.
The app is unstable.**

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