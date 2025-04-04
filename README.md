# Git-Go

A simple Git implementation written in Go. This project implements core Git functionality including repository initialization, object storage, tree creation, and commit management.

## Features

- `init`: Initialize a new Git repository
- `cat-file`: Display the contents of a Git object
- `hash-object`: Compute object ID and optionally creates a blob from a file
- `ls-tree`: List the contents of a tree object
- `write-tree`: Create a tree object from the current index
- `commit-tree`: Create a new commit object
- `clone`: Clone a remote repository

## Prerequisites

- Go 1.22 or later
- Bash shell (for running the provided script)

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd git-go
```

2. Make the run script executable:
```bash
chmod +x run.sh
```

## Usage

The project provides a wrapper script `run.sh` that builds and executes the Git implementation. All commands should be run through this script:

```bash
./run.sh <command> [arguments...]
```

### Available Commands

#### Initialize a Repository
```bash
./run.sh init
```

#### Create and Store a Blob
```bash
./run.sh hash-object -w <file>
```

#### Display Object Contents
```bash
./run.sh cat-file -p <sha>
```

#### List Tree Contents
```bash
./run.sh ls-tree --name-only <sha>
```

#### Create a Tree Object
```bash
./run.sh write-tree
```

#### Create a Commit
```bash
./run.sh commit-tree <tree-sha> -p <parent-sha> -m <message>
```

#### Clone a Repository
```bash
./run.sh clone <repository-url>
```