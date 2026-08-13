// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

/// @title TransactionLogger
/// @notice Audit log for completed RekberKuy escrow transactions on Avalanche.
/// @dev Gasless model: the backend Relayer bears every gas fee. This contract NEVER holds
///      funds, NEVER computes fees, and stores NO user PII — only keccak256 hashes plus the
///      amount are recorded, purely as event logs. See blockchain/README.md for the absolute
///      audit-log-only constraints.
contract TransactionLogger {
    /// @dev Offline admin wallet, sole authority to rotate the relayer. Immutable (stored in
    ///      code, no per-call SLOAD) — rotating the owner requires redeploy, by design, so the
    ///      admin key can live cold storage.
    address public immutable owner;

    /// @dev Authorized caller of logTransaction. Mutable so a compromised hot relayer key can
    ///      be rotated by the owner WITHOUT redeploying and splitting the audit trail across
    ///      two contracts. Reading it in onlyRelayer costs one (cold) SLOAD per call.
    address public relayer;

    /// @notice Emitted for every recorded transaction.
    /// @dev The ABI signature of this event AND of `logTransaction` MUST stay byte-identical
    ///      to `loggerABI` in apps/core-service/internal/relayer/relayer.go — changing either
    ///      breaks the Go Relayer's calldata packing / log decoding.
    ///      Indexing mirrors the Go ABI: txIdHash / buyerHash / sellerHash indexed for public
    ///      filtering; amount / timestamp non-indexed as event data.
    event TransactionLogged(
        bytes32 indexed txIdHash,
        uint256 amount,
        bytes32 indexed buyerHash,
        bytes32 indexed sellerHash,
        uint256 timestamp
    );

    /// @notice Emitted whenever the relayer changes (including the initial assignment from
    ///         address(0) in the constructor), so the full relayer history is publicly auditable.
    event RelayerUpdated(address indexed oldRelayer, address indexed newRelayer);

    error Unauthorized();
    error NotOwner();
    error ZeroOwnerAddress();
    error ZeroRelayerAddress();

    constructor(address _owner, address _relayer) {
        if (_owner == address(0)) revert ZeroOwnerAddress();
        if (_relayer == address(0)) revert ZeroRelayerAddress();
        owner = _owner;
        relayer = _relayer;
        emit RelayerUpdated(address(0), _relayer);
    }

    /// @notice Rotates the authorized relayer wallet. Owner-only.
    /// @dev Use to invalidate a compromised relayer key without redeploy. Emits RelayerUpdated
    ///      so the rotation is itself part of the public audit trail.
    function setRelayer(address newRelayer) external onlyOwner {
        if (newRelayer == address(0)) revert ZeroRelayerAddress();
        address old = relayer;
        relayer = newRelayer;
        emit RelayerUpdated(old, newRelayer);
    }

    /// @notice Records one completed transaction's audit entry. Relayer-only.
    /// @dev External value-type parameters are calldata by default (no `memory`).
    ///      `txIdHash` / `buyerHash` / `sellerHash` are keccak256 hashes produced off-chain by
    ///      the Relayer — raw identifiers never reach the chain.
    function logTransaction(
        bytes32 txIdHash,
        uint256 amount,
        bytes32 buyerHash,
        bytes32 sellerHash
    ) external onlyRelayer {
        emit TransactionLogged(txIdHash, amount, buyerHash, sellerHash, block.timestamp);
    }

    modifier onlyRelayer() {
        if (msg.sender != relayer) revert Unauthorized();
        _;
    }

    modifier onlyOwner() {
        if (msg.sender != owner) revert NotOwner();
        _;
    }
}
