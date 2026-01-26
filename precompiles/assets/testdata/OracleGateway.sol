// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "../IAssets.sol";

/// @title OracleGateway
/// @notice Minimal gateway for oracle-bridge e2e tests.
/// @dev Accepts oracle module calls and executes a deposit payload.
contract OracleGateway {
    address public immutable oracleCaller;

    event OracleReceived(uint32 srcChainId, uint64 nonce, bytes message);
    event DepositLST(uint32 srcChainId, bytes staker, bytes token, uint256 amount, bool success);

    constructor(address oracleCaller_) {
        require(oracleCaller_ != address(0), "oracle caller is zero");
        oracleCaller = oracleCaller_;
    }

    function oracleReceive(uint32 srcChainId, uint64 nonce, bytes calldata message) external {
        require(msg.sender == oracleCaller, "unauthorized caller");
        emit OracleReceived(srcChainId, nonce, message);

        require(message.length >= 97, "invalid message length");
        uint8 action = uint8(message[0]);
        bytes calldata payload = message[1:];

        if (action == 2) {
            // DepositLST payload: staker(32) | amount(32) | token(32)
            bytes calldata staker = payload[:32];
            uint256 amount = uint256(bytes32(payload[32:64]));
            bytes calldata token = payload[64:96];
            (bool success, ) = ASSETS_CONTRACT.depositLST(srcChainId, token, staker, amount);
            emit DepositLST(srcChainId, staker, token, amount, success);
            require(success, "deposit failed");
        } else {
            revert("unsupported action");
        }
    }
}
