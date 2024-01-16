# Build

## Code generation

```bash
go generate ./build/
```

## Container Image

# Git

The image uses the git command line tool to interact with github. Therefore, you
need to configure git to use your github account.

You can mount your .gitconfig file at /.gitconfig to at runtime.

Alternativly, you can use environment variables, if possible. For example
https://git-scm.com/book/en/v2/Git-Internals-Environment-Variables.

## SSH

If you want to use ssh, you need to create a key pair and add the public key to
your github account.

Additonally, you need to add the github host key to the known hosts file.

```bash
mkdir -p .ssh
ssh-keygen -t ed25519
ssh-keyscan -h github.com > .ssh/known_hosts
```

The .ssh folder will be mounted into the container at runtime.