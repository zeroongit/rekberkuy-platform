import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export default buildModule("TransactionLoggerModule", (m) => {
  // Production: pass DISTINCT wallets via --parameters params.json, e.g.
  //   { "owner": "0xOFFLINE_COLD...", "relayer": "0xRELAYER_HOT..." }
  // The owner should be a cold/offline key (it only rotates the relayer); the relayer is the
  // hot key the backend signs every logTransaction tx with (DEPLOYER_PRIVATE_KEY in
  // apps/core-service). Defaults to Hardhat account 0 (deployer) for local/dev deployment.
  const owner = m.getParameter("owner", m.getAccount(0));
  const relayer = m.getParameter("relayer", m.getAccount(0));

  const transactionLogger = m.contract("TransactionLogger", [owner, relayer]);

  return { transactionLogger };
});
