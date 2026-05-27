# Zero-OS v4 ![Tests](https://github.com/threefoldtech/zos4/workflows/Tests%20and%20Coverage/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/threefoldtech/zos)](https://goreportcard.com/report/github.com/threefoldtech/zos)

Zero-OS v4 is the next-generation autonomous operating system. It evolves the Zero-OS architecture with new capabilities, improved modularity, and enhanced performance.

## What this is

This repository contains the codebase for Zero-OS v4, the future direction of the autonomous node operating system. Zero-OS v4 builds on the principles of its predecessor — stateless operation, automated workload provisioning, and minimal administration — while introducing architectural improvements that support broader deployment scenarios and higher performance.

If you want to know about the history and decisions that motivated the creation of the V2 architecture, you can read [this article](docs/internals/history/readme.md).

## What this repository contains

- **Next-generation node runtime** — core services and daemons for V4
- **Modular provisioning subsystem** — workload deployment and lifecycle management
- **Networking stack** — overlay and direct networking support
- **Storage subsystem** — volume, cache, and persistent storage management
- **Identity, cryptography, and upgrade mechanisms**
- **Resource monitoring and capacity reporting**
- **Integration tests and development tooling** (QEMU-based testing environment)

## Role in the stack

Zero-OS v4 is the operating system layer for next-generation nodes. It provides the same foundational services as earlier Zero-OS versions — automated provisioning, networking, storage, and resource management — with an evolved architecture. It sits above the hardware and below user workloads, coordinating with grid infrastructure to receive reservations and report status. Zinit provides service supervision, and zosbase supplies shared primitives.

## ZOS / Zero-OS

ZOS, also known as Zero-OS, is the operating system layer used to run and manage nodes. It provides the low-level runtime environment for workloads, networking, storage, and automation.

## Mycelium

Mycelium is the network layer used to provide secure, peer-to-peer connectivity between nodes, services, and users. It enables decentralized networking across the infrastructure stack and is used as part of the ThreeFold Grid deployment.

## Relation to ThreeFold

This technology is used within the ThreeFold ecosystem and was first deployed on the ThreeFold Grid. The component itself is designed as reusable infrastructure technology and should be understood by its technical function first, independent of any specific deployment.

## Ownership

This repository is owned and maintained by TF-Tech NV, a Belgian company responsible for the development and maintenance of this technology.

## Documentation

Start exploring the codebase by first checking the [documentation](/docs) and [specification documents](/specs).

An [FAQ](./docs/faq/readme.md) is also available for common questions.

## Setting up your development environment

If you want to contribute, read the [contribution guideline](CONTRIBUTING.md) and the documentation to set up your [development environment](qemu/README.md).

## Grid Networks

Zero-OS is deployed on several network environments:

- **production network**: Released stable versions. Used to run the real grid. Cannot be reset. Only stable and battle-tested features reach this level. [Dashboard](https://dashboard.grid.tf/)
- **test network**: Mostly stable features that need to be tested at scale. Can be reset occasionally. [Dashboard](https://dashboard.test.grid.tf/)
- **QA network**: Internal testing of new features. Can be behind development. [Dashboard](https://dashboard.qa.grid.tf/)
- **dev network**: Ephemeral network for developing and testing new features. Can be created and reset at any time. [Dashboard](https://dashboard.dev.grid.tf/)

Learn more about the different networks by reading the [upgrade documentation](/docs/internals/identity/upgrade.md#philosophy).

### Provisioning of workloads

Zero-OS does not expose an interface. Instead, it waits for reservations to happen on a trusted source, and once a reservation is available, the node applies it to reality. You can start reading about [provisioning](./docs/provision) in this document.

## Community

If you have questions or want to connect, you can find the community on:

- Telegram: <https://t.me/zero_os_tech>
- Matrix: #zero-os:matrix.org

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
Copyright (c) TF-Tech NV.
