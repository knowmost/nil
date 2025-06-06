import type { Abi } from "abitype";
import { task } from "hardhat/config";
import {
    ProcessedReceipt,
} from "@nilfoundation/niljs";
import { loadNilSmartAccount } from "./nil-smart-account";
import { L2NetworkConfig, loadNilNetworkConfig } from "../deploy/config/config-helper";
import { encodeFunctionData } from "viem";

// npx hardhat relay-l2-message --networkname local
task("relay-l2-message", "relay the l1-rbidge event data as message on to L2BridgeMessenger from Nil Chain")
    .addParam("networkname", "The network to use") // Mandatory parameter
    .setAction(async (taskArgs) => {

        // Dynamically load artifacts
        const L2BridgeMessengerJson = await import("../artifacts/contracts/bridge/l2/L2BridgeMessenger.sol/L2BridgeMessenger.json");
        if (!L2BridgeMessengerJson || !L2BridgeMessengerJson.default || !L2BridgeMessengerJson.default.abi || !L2BridgeMessengerJson.default.bytecode) {
            throw Error(`Invalid L2BridgeMessengerJson ABI`);
        }

        const networkName = taskArgs.networkname;
        const deployerAccount = await loadNilSmartAccount();

        if (!deployerAccount) {
            throw Error(`Invalid Deployer SmartAccount`);
        }

        const balance = await deployerAccount.getBalance();

        if (!(balance > BigInt(0))) {
            throw Error(`Insufficient or Zero balance for smart-account: ${deployerAccount.address}`);
        }

        // save the L2BridgeMessenger Address in the json config for l2
        const l2NetworkConfig: L2NetworkConfig = loadNilNetworkConfig(networkName);

        const messageSender = l2NetworkConfig.l2TestConfig.messageSentEvent.messageSender;
        const messageTarget = l2NetworkConfig.l2TestConfig.messageSentEvent.messageTarget;
        const messageType = l2NetworkConfig.l2TestConfig.messageSentEvent.messageType;
        const messageNonce = l2NetworkConfig.l2TestConfig.messageSentEvent.messageNonce;
        const message = l2NetworkConfig.l2TestConfig.messageSentEvent.message;
        const messageExpiryTime = l2NetworkConfig.l2TestConfig.messageSentEvent.messageExpiryTime;
        const feeCredit = l2NetworkConfig.l2TestConfig.messageSentEvent.feeCreditData.feeCredit;
        const maxFeePerGas = l2NetworkConfig.l2TestConfig.messageSentEvent.feeCreditData.maxFeePerGas;
        const maxPriorityFeePerGas = l2NetworkConfig.l2TestConfig.messageSentEvent.feeCreditData.maxPriorityFeePerGas;

        console.log(`EventData to be relayed is: \n `);
        console.log(`Message Sender: ${messageSender}`);
        console.log(`Message Target: ${messageTarget}`);
        console.log(`Message Type: ${messageType}`);
        console.log(`Message Nonce: ${messageNonce}`);
        console.log(`Message: ${message}`);
        console.log(`Message Expiry Time: ${messageExpiryTime}`);
        console.log(`Fee Credit: ${feeCredit}`);
        console.log(`Max Fee Per Gas: ${maxFeePerGas}`);
        console.log(`Max Priority Fee Per Gas: ${maxPriorityFeePerGas}`);

        const relayMessage = encodeFunctionData({
            abi: L2BridgeMessengerJson.default.abi as Abi,
            functionName: "relayMessage",
            args: [messageSender, messageTarget, messageType, messageNonce, message, messageExpiryTime],
        });

        const relayEthDepositMessageResponse = await deployerAccount.sendTransaction({
            to: l2NetworkConfig.l2BridgeMessengerConfig.l2BridgeMessengerContracts.l2BridgeMessengerProxy as `0x${string}`,
            data: relayMessage,
            feeCredit: BigInt(feeCredit),
            maxFeePerGas: BigInt(maxFeePerGas),
            maxPriorityFeePerGas: BigInt(maxPriorityFeePerGas)
        });

        const relayEthDepositMessageTxnReceipts: ProcessedReceipt[] = await relayEthDepositMessageResponse.wait();

        const relayedEthDepositMessageTxnReceipt: ProcessedReceipt = relayEthDepositMessageTxnReceipts[0] as ProcessedReceipt;

        const outputReceipts: ProcessedReceipt[] = relayedEthDepositMessageTxnReceipt.outputReceipts as ProcessedReceipt[];

        console.log(`outputReceipt is: ${JSON.stringify(outputReceipts, bigIntReplacer, 2)}`)

        const outputReceipt: ProcessedReceipt = outputReceipts[0] as ProcessedReceipt;

        console.log(`outputReceipt extracted as: ${JSON.stringify(outputReceipt, bigIntReplacer, 2)}`);

        // check the first element in the ProcessedReceipt and verify if it is successful
        if (!outputReceipt.success) {
            throw Error(`Failed to relay message
            on the L2BridgeMessenger contract: ${l2NetworkConfig.l2BridgeMessengerConfig.l2BridgeMessengerContracts.l2BridgeMessengerProxy}`);
        }

        console.log(`successfully relayed EthDepositMessage on to L2BridgeMessenger with transactionReceipt: ${JSON.stringify(relayEthDepositMessageTxnReceipts[0], bigIntReplacer, 2)}`);
    });

export function bigIntReplacer(unusedKey: string, value: unknown): unknown {
    return typeof value === "bigint" ? value.toString() : value;
}