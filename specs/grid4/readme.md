# Grid v4 Documentation

## Introduction

This document provides an overview of tf-grid-v4, focusing on how nodes operate and communicate with each other outside of blockchain interactions. Grid v4 is the evolution of Grid3, with significant changes in how nodes register and communicate.

## Definitions

- **Node**: A machine that runs ZOS (Zero-OS) operating system.
- **Registrar**: An HTTP server that handles node registration and version control.
- **Twin**: A digital representation of a node or user on the grid, associated with a public/private key pair.
- **Farm**: A collection of nodes operated by a single entity.

## Key Differences Between Grid3 and Grid4

| Feature | Grid3 | Grid4 |
|---------|-------|-------|
| Registration | Nodes register on blockchain (grid-db) | Nodes register on HTTP server (registrar) |
| Node-Registrar Communication | Uses RMB (Reliable Message Bus) | Uses signed HTTP requests |
| Node-to-Node Communication | Uses RMB (Reliable Message Bus) | Not implemented yet (planned to use Open RPC API) |
| Version Control | Through blockchain | Through registrar server |
| Contract Management | On blockchain | Not fully implemented yet |

## Overview

Grid v4 operates with the following key components:

1. **Node Identity**: Each node has a unique identity based on a public/private key pair.
2. **Registration**: Nodes register with the registrar server through signed HTTP requests.
3. **Version Control**: The registrar server manages version control for nodes.
4. **Workload Deployment**: Users can deploy workloads on nodes.

## Node Registration Process

When a node boots for the first time, it follows these steps:

1. The `identityd` daemon generates or loads a key pair that represents the node's identity.
   - This key pair is used for signing all communications with the registrar server.
   - The public key serves as the node's unique identifier.

2. The `registrar_light` module collects node information:
   - Hardware capacity (CPU, memory, storage, GPU)
   - Geographic location (obtained via geoip service)
   - Network interfaces (name, MAC address, IPs)
   - Hardware details (secure boot status, virtualization status, serial number)

3. The node sends this information to the registrar server through signed HTTP requests via the `registrar_gateway`:
   - The request includes an authentication header with the signature.
   - The registrar server validates the signature using the node's public key.

4. Account creation and management:
   - If this is the first time the node connects, it creates a new account (twin) on the registrar.
   - If the node already has an account, it ensures the account is active.
   - The registrar assigns a twin ID to the node.

5. Node registration:
   - The node registers itself with its farm ID, twin ID, resources, location, and interfaces.
   - The registrar assigns a node ID to the node.
   - If the node was previously registered, it updates its information.

6. Periodic updates:
   - The node periodically checks its account status (every 30 minutes).
   - The node updates its information on the registrar server (every 24 hours).
   - The node updates uptime on the registrar server (every 40 hours).
   - If the node's network address changes, it immediately re-registers.

## Node Architecture

The node runs several core modules that work together:

1. **identityd**: Manages node identity and cryptographic operations.
2. **registrar_light**: Handles node registration with the registrar server.
3. **noded**: Reports node resources and monitors system health.
4. **provisiond**: Manages workload deployments.
5. **storaged**: Handles disk and volume management.
6. **netlightd**: Manages network resources.
7. **vmd**: Manages virtual machines.
8. **contd**: Handles container deployments.
9. **flistd**: Manages file system mounts.
10. **powerd**: Manages power state.

## Version Control and Upgrades

Grid v4 implements a sophisticated version control system:

1. **Version Management via Registrar**:
   - The registrar server maintains the current approved version of ZOS.
   - The registrar provides a `ZosVersion` object that includes:
     - The current approved version string
     - A `SafeToUpgrade` flag that controls rollout
     - A list of test farms for A/B testing

2. **Update Detection**:
   - The `upgrade` package in the node checks for updates periodically (every 60 minutes with jitter).
   - The node compares its current version with the version from the registrar.
   - If versions differ and the `SafeToUpgrade` flag is true (or the node is on a test farm), the update process begins.

3. **Update Process**:
   - Updates are fetched from a hub server as flist packages.
   - The node first downloads and installs all dependency packages.
   - The ZOS package itself is updated last to ensure all dependencies are in place.
   - The update process is atomic - either all packages are updated successfully, or none are.

4. **Safe Update Mechanism**:
   - The node uses a multi-stage bootstrap process to ensure reliable updates.
   - Updates are applied in a way that prevents interruption during critical operations.
   - If an update fails, the node can roll back to the previous working state.
   - The registrar can enable A/B testing by setting specific farms as test farms.
   - Test farms receive updates first, allowing for validation before wider deployment.
   - The `SafeToUpgrade` flag controls whether non-test farms should update.
   - This allows for gradual rollout of updates across the grid.

## Workload Provisioning

The provisioning system in Grid v4 handles the deployment of workloads

This is not fully implemented yet

1. **Provisioning Engine**:
   - The `provision` engine (`provisiond` module) manages the lifecycle of all workloads.
   - It uses a queue-based system to process workload operations in the correct order.
   - The engine maintains a persistent storage of all deployments and their states.

2. **Workload Types**:
   - Workloads are defined as deployments with specific types:
     - `ZMountType`: File system mounts
     - `VolumeType`: Storage volumes
     - `QuantumSafeFSType`: Secure file systems
     - `NetworkLightType`: Network configurations
     - `PublicIPv4Type`: Public IPv4 addresses
     - `ZMachineLightType`: Virtual machines
     - `ZLogsType`: Log forwarding

3. **Workload Operations**:
   - The engine processes several types of operations:
     - `Provision`: Deploy a new workload
     - `Deprovision`: Remove an existing workload
     - `Update`: Modify an existing workload
     - `Pause`: Temporarily suspend a workload
     - `Resume`: Reactivate a suspended workload

4. **Type Managers**:
   - Each workload type has a dedicated manager that implements:
     - `Provision`: Create the workload resources
     - `Deprovision`: Clean up the workload resources
     - Optional `Update`: Modify the workload without recreating it
     - Optional `Pause`/`Resume`: Suspend/resume the workload

5. **Deployment Processing**:
   - Workloads are processed in a specific order to ensure dependencies are met:
     - Storage volumes are created first
     - Networks are configured next
     - VMs and containers are deployed last
   - This ordering ensures that resources required by VMs/containers are available when needed.

6. **State Management**:
   - The provisioning system maintains the state of all deployments:
     - `StateOk`: Workload is running correctly
     - `StateError`: Workload deployment failed
     - `StatePaused`: Workload is temporarily suspended
     - `StateDeleted`: Workload has been removed
   - Each workload result includes the state, creation timestamp, and any error messages.

## Communication Flow

1. **Client Interaction**:
   - Users interact with the grid through client tools or APIs.
   - The primary interface is zos-api-light.

2. **Request Authentication**:
   - All requests are signed using the sender's private key.
   - The signature is included in the `X-Auth` HTTP header.
   - This ensures that only authorized users can interact with the grid.

3. **Node-Registrar Communication**:
   - Nodes communicate with the registrar server through signed HTTP requests.
   - The registrar validates the signature using the node's public key.
   - This replaces the RMB (Reliable Message Bus) used in Grid3 for registration.

4. **Node-to-Node Communication**:
   - Direct node-to-node communication is not yet implemented in Grid v4.
   - Future versions will implement an Open RPC API for node-to-node communication.
   - This will replace the RMB used in Grid3 for peer communication.

5. **Inter-Module Communication**:
   - Within a node, modules communicate through a message bus (zbus) using Redis.
   - This provides a reliable and efficient way for modules to interact.
   - Each module exposes a set of methods that can be called by other modules.

6. **Status Reporting**:
   - Nodes periodically report their status to the registrar server.
   - This includes uptime, resource usage, and health information.
   - The registrar uses this information to maintain an up-to-date view of the grid.

## Security

Grid v4 implements several security measures:

1. **Cryptographic Identity**:
   - All entities (nodes, users) have a unique identity based on Ed25519/SR25519 key pairs.
   - The public key serves as the identity, while the private key is used for signing.
   - This provides a secure and verifiable way to identify entities.

2. **Signed Communication**:
   - All communication is signed using Ed25519/SR25519 cryptographic signatures.
   - This ensures that messages cannot be tampered with or forged.
   - The signature is verified by the recipient before processing the request.

3. **Authentication and Authorization**:
   - The registrar server validates node authenticity using signatures.
   - Workload deployments are authorized based on twin identity.
   - Only authorized users can deploy workloads on nodes.

## Future Developments

While Grid v4 is operational, some components are still under development:

1. **Node-to-Node Communication**:
   - Direct node-to-node communication is not yet implemented.
   - A new Open RPC API will be developed to replace the RMB used in Grid3.
   - This will enable peer-to-peer communication between nodes for distributed workloads.

2. **Contract Management**:
   - Contract management is not fully implemented yet.
   - This will provide a way to manage agreements between users and nodes.
   - It will include billing, resource allocation, and service level agreements.
