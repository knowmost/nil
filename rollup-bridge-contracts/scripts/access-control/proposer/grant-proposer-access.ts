import { Contract } from 'ethers';
import * as fs from 'fs';
import * as path from 'path';
import {
    loadL1NetworkConfig,
    isValidAddress,
} from '../../../deploy/config/config-helper';
import { getAllProposers } from './get-all-proposers';

const abiPath = path.join(
    __dirname,
    '../../../artifacts/contracts/interfaces/INilAccessControlUpgradeable.sol/INilAccessControlUpgradeable.json',
);
const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8')).abi;

// npx hardhat run scripts/access-control/proposer/grant-proposer-access.ts --network sepolia
// Function to grant proposer access
export async function grantProposerAccess(proposerAddress: string) {
    // Lazy import inside the function
    // @ts-ignore
    const { ethers, network } = await import('hardhat');
    const networkName = network.name;
    const config = loadL1NetworkConfig(networkName);

    if (!isValidAddress(config.nilRollup.nilRollupContracts.nilRollupProxy)) {
        throw new Error('Invalid nilRollupProxy address in config');
    }

    const [signer] = await ethers.getSigners();

    const nilRollupInstance = new ethers.Contract(
        config.nilRollup.nilRollupContracts.nilRollupProxy,
        abi,
        signer,
    ) as Contract;

    const tx = await nilRollupInstance.grantProposerAccess(proposerAddress);
    await tx.wait();

    console.log(`Proposer access granted to ${proposerAddress}`);
}

// Main function to call the grantProposerAccess function
async function main() {
    const proposerAddress = '0x7A2f4530b5901AD1547AE892Bafe54c5201D1206';
    await grantProposerAccess(proposerAddress);
}

main().catch((error) => {
    console.error(error);
    process.exit(1);
});
