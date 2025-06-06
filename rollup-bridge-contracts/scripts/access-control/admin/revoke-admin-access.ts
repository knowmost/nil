import { Contract } from 'ethers';
import * as fs from 'fs';
import * as path from 'path';
import { getRoleMembers } from '../get-role-members';
import { DEFAULT_ADMIN_ROLE } from '../../utils/roles';
import { isAnAdmin } from './is-an-admin';
import { isValidAddress, loadL1NetworkConfig } from '../../../deploy/config/config-helper';

// Load the ABI from the JSON file
const abiPath = path.join(
    __dirname,
    '../../../artifacts/contracts/interfaces/INilAccessControl.sol/INilAccessControl.json',
);
const abi = JSON.parse(fs.readFileSync(abiPath, 'utf8')).abi;

// npx hardhat run scripts/access-control/admin/revoke-admin-access.ts --network sepolia
// Function to revoke-admin-access
export async function revokeAdminAccess(adminAddress: string) {
    // Lazy import inside the function
    // @ts-ignore
    const { ethers, network } = await import('hardhat');
    const networkName = network.name;
    const config = loadL1NetworkConfig(networkName);

    // Validate configuration parameters
    if (!isValidAddress(config.nilRollup.nilRollupContracts.nilRollupProxy)) {
        throw new Error('Invalid nilRollupProxy address in config');
    }

    // Get the signer (default account)
    const [signer] = await ethers.getSigners();

    let isAnAdminResponse: Boolean = await isAnAdmin(adminAddress);

    if (!isAnAdminResponse) {
        throw new Error(
            `account: ${adminAddress} doesnot have admin-role-access. \n cannot revoke admin access from the account which is not an existing admin.`,
        );
    }

    // Create a contract instance
    const nilRollupInstance = new ethers.Contract(
        config.nilRollup.nilRollupContracts.nilRollupProxy,
        abi,
        signer,
    ) as Contract;

    // Grant proposer access
    const tx = await nilRollupInstance.removeAdmin(adminAddress);
    await tx.wait();

    console.log(`admin access revoked from ${adminAddress}`);

    isAnAdminResponse = await isAnAdmin(adminAddress);

    if (isAnAdminResponse) {
        throw new Error(
            `account: ${adminAddress} still have admin-role-access. \n revoke admin access from the account is not successful.`,
        );
    }
}

// Main function to call the grantAdminAccess function
async function main() {
    const adminAddress = '0x7A2f4530b5901AD1547AE892Bafe54c5201D1206';
    await revokeAdminAccess(adminAddress);
}

main().catch((error) => {
    console.error(error);
    process.exit(1);
});
