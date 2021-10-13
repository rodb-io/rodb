---
layout: default
title: Get started
permalink: /
---

# What is RODB?

RODB is a fast and configurable tool to create micro-services from static data files.

It currently supports input files as CSV, XML and JSON, and can create JSON APIs from it.

It also handles parsing and indexing the data.

# Architecture

TODO architecture and components

# Getting started

TODO provide multiple ways: bash, docker-compose, kubernetes...
TODO mention examples

# Configuration file structure

The configuration file allows to set-up each layer separately.
Every component has a `name` and a `type`.

The name must be unique among each layer (you can not have two parsers named `foo`, but you can have a `foo` parser and a `foo` index).

You can find all the available component types and configuration settings in the [documentation](/configuration/inputs).

```yaml
parsers:
  - name: someParser
    type: integer
    ...
  - name: anotherParser
    type: float
    ...
inputs:
  - name: someInput
    type: xml
    ...
  - name: anotherInput
    type: csv
    ...
indexes:
  - name: someIndex
    type: sqlite
    ...
  - name: anotherIndex
    type: fst5
    ...
outputs:
  - name: someOutput
    type: jsonArray
    ...
  - name: anotherInput
    type: jsonObject
    ...
services:
  - name: someService
    type: http
    ...
  - name: anotherService
    type: http
    ...
```

The configuration file structure is also provided as a JSON-schema [here](https://github.com/rodb-io/rodb/blob/master/docs/schema/config.yaml).

# Command line flags

The following arguments are available:
- `--config`, `-c`: Custom path to the configuration file. The default value is `rodb.yaml` (in the current working directory).
- `--loglevel`, `-l`: Changes the logging level. Supported values: `panic`, `fatal`, `error`, `warn[ing]`, `info`, `debug`, `trace`. The default value is `info`.
