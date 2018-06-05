cherami-server [![Build Status](https://travis-ci.org/uber/cherami-server.svg?branch=master)](https://travis-ci.org/uber/cherami-server) [![Coverage Status](https://coveralls.io/repos/uber/cherami-server/badge.svg?branch=master&service=github)](https://coveralls.io/github/uber/cherami-server?branch=master)
==============
[Cherami](https://eng.uber.com/cherami) is a distributed, scalable, durable, and highly available message queue system we developed at Uber Engineering to transport asynchronous tasks. 

This repo contains the source code of Cherami server, cross-zone replicator server, and several tools. Your application needs to use the client to interact with the server. The client can be found [here](https://github.com/uber/cherami-client-go).

Getting started
---------------

Using:
Go version 1.8 [`issue with newer Go version`](https://github.com/uber/cherami-server/issues/350)
`go version go1.8 linux/amd64`

Built on Ubuntu:
'Ubuntu 14.04.5 LTS'

GCC version:
`gcc version 4.8.4 (Ubuntu 4.8.4-2ubuntu1~14.04.3)` [`issue with GCC 4.8.4 and stdatomic.h`](https://github.com/uber/cherami-server/issues/350)

![alt text](assets/stdatomic_error.png "stdatomic.h problem with GCC 4.8.4")

Remove the following lines in order to get it to build, which I have and it is now running..
from jemalloc_internal.h
in:
'vendor\github.com\cockroachdb\c-jemalloc\linux_includes\internal\include\jemalloc\internal'

removed(lines 145-147):
#ifdef JEMALLOC_C11ATOMICS
#include <stdatomic.h>
#endif

reference: https://stackoverflow.com/questions/20326604/stdatomic-h-in-gcc-4-8

To get cherami-server:

```
git clone git@github.com:valdezm/cherami-server.git $GOPATH/src/github.com/valdezm/cherami-server
```

Build
-----
We use [`glide`](https://glide.sh) to manage Go dependencies. Please make sure `glide` is in your PATH before you attempt to build.

* Build the `cherami-server` and other binaries (will not run test):
```
make bins
```

Local Test
----------
We need a Cassandra running locally in order to run the integration tests. Please make sure `cqlsh` is in `/usr/local/bin`, and it can connect to the local Cassandra server.
```
make test
```

Run Cherami locally
-------------------
* Setup the cherami keyspace for metadata:
```
RF=1 ./scripts/cherami-setup-schema
```

* The service can be started as follows:
```
CHERAMI_ENVIRONMENT=local ./cherami-server start all
```

Note: `cherami-server` is configured via `config/base.yaml` with some parameters overriden by `config/local.yaml`. In this config, Cherami is bound to `localhost`.

One can use the CLI to verify if Cherami is running properly:
```
./cherami-cli --env=prod --hostport=127.0.0.1:4922 create destination /test/cherami
```


Cherami keyspace
----------------
I dumped the keyspace to this file: [`cherami_cql.txt`](assets/cherami_cql.txt)



Deploy Cherami as a cluster
---------------------------
Documentation coming soon....

Contributing
------------

We'd love your help in making Cherami great. If you find a bug or need a new feature, open an issue and we will respond as fast as we can. If you want to implement new feature(s) and/or fix bug(s) yourself, open a pull request with the appropriate unit tests and we will merge it after review.

**Note:** All contributors also need to fill out the [Uber Contributor License Agreement](http://t.uber.com/cla) before we can merge in any of your changes.

Documentation
--------------

Interested in learning more about Cherami? Read the blog post:
[eng.uber.com/cherami](https://eng.uber.com/cherami/)

License
-------
MIT License, please see [LICENSE](https://github.com/valdezm/cherami-server/blob/master/LICENSE) for details.
