import { hasRole } from '../has-a-role';
import { DEFAULT_ADMIN_ROLE } from '../../utils/roles';

// Function to check if an address is a Admin
export async function isAnAdmin(account: string): Promise<Boolean> {
    const isAnAdminResponse = await hasRole(DEFAULT_ADMIN_ROLE, account);

    // Convert the response to a boolean
    const isAnAdminIndicator = Boolean(isAnAdminResponse);

    return isAnAdminIndicator;
}
