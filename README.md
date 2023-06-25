## Restaking Protocol

Restaking Protocol is a solution built on Cosmos SDK and IBC-go to address the security concerns in newly launched Proof-of-Stake (POS) chains, specifically tackling the problem of insufficient staked funds leading to network vulnerability.

### Problem Statement

In new POS chains, the limited number of validators or low amount of staked funds can pose security risks and make the network susceptible to attacks. Insufficient staked funds compromise the network's integrity and stability, as validators lack the necessary collateral to participate actively in the consensus and validation process.

### Solution

Restaking Protocol introduces a mechanism to encourage restaking and increase the amount of staked funds in the network. By incentivizing validators to restake their funds, the protocol aims to strengthen the network's security and resilience against potential attacks.

The protocol achieves this by offering rewards, delegation incentives, or other benefits to validators who actively participate in the consensus process and keep their funds staked. By promoting restaking, more validators are attracted to the network, ensuring a larger pool of staked funds and enhancing overall security.[Specific implementation](https://github.com/celinium-network/restaking_protocol/tree/main/x/restaking/README.md)

### Repository Structure

This repository contains the implementation of Restaking Protocol using the following technologies:

- Cosmos SDK: A development framework for building blockchain applications, providing the necessary tools and modules.
- IBC-go (Inter-Blockchain Communication): A module within the Cosmos ecosystem that enables interoperability and cross-chain communication between blockchains.

Restaking Protocol consists of several main modules that play crucial roles in its functionality. Here are the descriptions of four key modules:
1. epochs: Timer of blockchain.
2. multistaking: The multistaking module expands the staking module to enable the participation of multiple tokens in POS staking. It allows for the staking of multiple tokens and relies on an oracle to calculate the conversion between the staked tokens and the native tokens of the blockchain.
3. restaking/consumer: The module is designed for chains that wish to introduce restaking functionality
4. restaking/coordinator: The module is designed for the chain that aims to provide restaking service.

### Contributing

Contributions to Restaking Protocol are welcome. If you encounter any issues or have suggestions for improvements, please open an issue or submit a pull request.

### License

This repository is licensed under the [Apache License](LICENSE).