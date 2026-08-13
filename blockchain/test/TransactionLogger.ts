import assert from "node:assert/strict";
import { describe, it } from "node:test";

import { getAddress } from "viem";

import { network } from "hardhat";

// Mirrors apps/core-service/internal/relayer/relayer.go: only keccak256 HASHES
// (txIdHash / buyerHash / sellerHash) + the amount are ever sent on-chain.
// No raw identifiers, no PII — see blockchain/README.md audit-log-only constraints.
describe("TransactionLogger", async function () {
  const { viem } = await network.create();
  const publicClient = await viem.getPublicClient();
  const [owner, relayer, other, newRelayer] = await viem.getWalletClients();

  const txIdHash = "0x1111111111111111111111111111111111111111111111111111111111111111" as const;
  const buyerHash = "0x2222222222222222222222222222222222222222222222222222222222222222" as const;
  const sellerHash = "0x3333333333333333333333331111111111111111111111111111111111113333" as const;
  const amount = 500_000n;
  const ZERO_ADDRESS = "0x0000000000000000000000000000000000000000" as const;

  const deploy = (ownerAddr: string, relayerAddr: string) =>
    viem.deployContract("TransactionLogger", [ownerAddr, relayerAddr]);

  // logTransaction must be sent FROM the current relayer identity, so use an explicit
  // wallet client rather than the contract's default account.
  const logFrom = (client: typeof relayer, logger: Awaited<ReturnType<typeof deploy>>) =>
    client.writeContract({
      address: logger.address,
      abi: logger.abi,
      functionName: "logTransaction",
      args: [txIdHash, amount, buyerHash, sellerHash],
    });

  it("reverts when deployed with the zero address as the owner", async function () {
    await assert.rejects(deploy(ZERO_ADDRESS, relayer.account.address));
  });

  it("reverts when deployed with the zero address as the relayer", async function () {
    await assert.rejects(deploy(owner.account.address, ZERO_ADDRESS));
  });

  it("emits TransactionLogged with the exact args when called by the relayer", async function () {
    const logger = await deploy(owner.account.address, relayer.account.address);

    const txHash = await logFrom(relayer, logger);
    const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
    const block = await publicClient.getBlock({ blockNumber: receipt.blockNumber });

    const events = await publicClient.getContractEvents({
      address: logger.address,
      abi: logger.abi,
      eventName: "TransactionLogged",
      fromBlock: receipt.blockNumber,
      toBlock: receipt.blockNumber,
    });

    assert.equal(events.length, 1, "exactly one TransactionLogged event expected");
    const args = events[0].args;
    assert.equal(args.txIdHash, txIdHash);
    assert.equal(args.amount, amount);
    assert.equal(args.buyerHash, buyerHash);
    assert.equal(args.sellerHash, sellerHash);
    assert.equal(args.timestamp, block.timestamp, "timestamp must equal block.timestamp");

    // The backend Relayer bears this gas on EVERY logged transaction (operational cost).
    console.log(
      `    gas: logTransaction() call used ${receipt.gasUsed.toString()} gas`,
    );
  });

  it("reverts when a non-relayer (arbitrary address) calls logTransaction", async function () {
    const logger = await deploy(owner.account.address, relayer.account.address);
    await assert.rejects(logFrom(other, logger), /unauthorized|reverted/i);
  });

  it("emits RelayerUpdated for the initial relayer in the constructor", async function () {
    const logger = await deploy(owner.account.address, relayer.account.address);

    const events = await publicClient.getContractEvents({
      address: logger.address,
      abi: logger.abi,
      eventName: "RelayerUpdated",
    });

    assert.equal(events.length, 1, "constructor must emit RelayerUpdated once");
    assert.equal(getAddress(events[0].args.oldRelayer), ZERO_ADDRESS);
    assert.equal(getAddress(events[0].args.newRelayer), getAddress(relayer.account.address));
  });

  it("lets the owner rotate the relayer and emits RelayerUpdated", async function () {
    const logger = await deploy(owner.account.address, relayer.account.address);

    const txHash = await owner.writeContract({
      address: logger.address,
      abi: logger.abi,
      functionName: "setRelayer",
      args: [newRelayer.account.address],
    });
    const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });

    const events = await publicClient.getContractEvents({
      address: logger.address,
      abi: logger.abi,
      eventName: "RelayerUpdated",
      fromBlock: receipt.blockNumber,
      toBlock: receipt.blockNumber,
    });

    assert.equal(events.length, 1, "setRelayer must emit RelayerUpdated once");
    assert.equal(getAddress(events[0].args.oldRelayer), getAddress(relayer.account.address));
    assert.equal(getAddress(events[0].args.newRelayer), getAddress(newRelayer.account.address));
    assert.equal(getAddress(await logger.read.relayer()), getAddress(newRelayer.account.address));
  });

  it("reverts when a non-owner calls setRelayer", async function () {
    const logger = await deploy(owner.account.address, relayer.account.address);

    await assert.rejects(
      other.writeContract({
        address: logger.address,
        abi: logger.abi,
        functionName: "setRelayer",
        args: [newRelayer.account.address],
      }),
      /notowner|reverted/i,
    );
  });

  it("reverts when setRelayer is called with the zero address", async function () {
    const logger = await deploy(owner.account.address, relayer.account.address);

    await assert.rejects(
      owner.writeContract({
        address: logger.address,
        abi: logger.abi,
        functionName: "setRelayer",
        args: [ZERO_ADDRESS],
      }),
      /zerorelayeraddress|reverted/i,
    );
  });

  it("blocks the old relayer and allows the new relayer after rotation", async function () {
    const logger = await deploy(owner.account.address, relayer.account.address);

    await owner.writeContract({
      address: logger.address,
      abi: logger.abi,
      functionName: "setRelayer",
      args: [newRelayer.account.address],
    });

    // old relayer can no longer log
    await assert.rejects(logFrom(relayer, logger), /unauthorized|reverted/i);

    // new relayer can log
    const txHash = await logFrom(newRelayer, logger);
    const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
    assert.ok(receipt.status === "success", "new relayer's logTransaction must succeed");
  });
});
