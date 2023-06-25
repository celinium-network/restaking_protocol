# Restaking
## Initialization
The Restaking protocol is built on top of the IBC (Inter-Blockchain Communication) protocol. Its initialization process is closely tied to the IBC handshake. The steps for initialization are as follows:

1. The Consumer initiates a proposal on the Coordinator to add a Consumer, which includes the metadata of the Consumer Chain.
2. Once the proposal is approved, the Consumer initiates an IBC handshake.
3. The Coordinator verifies the handshake information against the proposal.
4. The Consumer sends the current Validator Set to the Coordinator.
5. The Coordinator saves the Validator Set of the Consumer.
6. Handshake is successfully completed.

## Consumer Validator Set Change
The Coordinator needs to track the creation and removal of Validators on the Consumer. Since there may be a large number of Validators on-chain, during initialization, we only select and send the Validators with block voting power. Subsequently, by tracking the creation and removal of Validators on the Consumer, we can ensure a fair opportunity for each validator to receive restaking funds. 

## Operator
The Operator is a core role within the Restaking Protocol and is closely related to delegation, undelegation, and slashing. The Operator exists on the Coordinator. [Operator definition](https://github.com/celinium-network/restaking_protocol/blob/main/x/restaking/coordinator/types/coordinator.pb.go#L126). The Operator defines the token for restaking, records the amount of restaking, the total shares, and the validator on the consumer (each consumer can only select one validator).

Creating an Operator may require staking some funds and receiving a portion of stake rewards as an incentive. These specifics will be determined as the economic model matures. The Owner of the Operator has the authority to maintain the list of consumer validators, which may be adjusted through governance.

### Delegate
Users delegate funds to the Operator and receive shares in return. These shares are used for redemption and reward distribution. The Operator periodically notifies its consumer of the delegated funds it has received. Upon receiving this notification, the consumer informs the Multistaking module to stake the non-native tokens. Currently, Multistaking works by calculating and minting an equivalent amount of native tokens using an oracle. These minted tokens are then used for staking. The Multistaking mechanism is customizable, allowing consumers to adjust it based on their specific business requirements.

### Undelegate
When a user initiates an undelegation, their shares are burned. The Operator periodically consolidates the undelegated funds and notifies the consumer. The related delegation on the consumer side should be immediately reduced, and the Coordinator is informed of the unbonding period. On the Coordinator, the Operator generates an UnbondingEntry based on the consumer's latest unbonding time. The UnbondingEntry operates similarly to the staking mechanism in the Cosmos SDK. It is a time-ordered queue, and when the UnbondingEntry's complete time is reached, the UnbondingEntry is destroyed, and the user receives their funds.

### Slash
When a validator on the consumer side is slashed, the Restaking Consumer module receives a notification. The Restaking Consumer module sends the slash information, along with all the operators delegating to that validator, to the Coordinator. Upon receiving the slash information, the Coordinator immediately burns the corresponding tokens and notifies other consumers to modify their delegation corresponding to the Operator-Validator pair.

## 