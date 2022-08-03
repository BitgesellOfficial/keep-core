import { task, types } from "hardhat/config"

import {
  calculateTokensNeededForStake,
  mint,
  stake,
  authorize,
  register,
} from "./utils"

import type { BigNumberish } from "ethers"
import type { HardhatRuntimeEnvironment } from "hardhat/types"

export const TASK_INITIALIZE = "initialize"
export const TASK_MINT = `${TASK_INITIALIZE}:mint`
export const TASK_STAKE = `${TASK_INITIALIZE}:stake`
export const TASK_AUTHORIZE = `${TASK_INITIALIZE}:authorize`
export const TASK_REGISTER = `${TASK_INITIALIZE}:register`
const TASK_AUTHORIZE_BEACON = `${TASK_AUTHORIZE}:beacon`
const TASK_REGISTER_BEACON = `${TASK_REGISTER}:beacon`

task(TASK_INITIALIZE, "Initializes staking for an operator")
  .addParam("owner", "Stake Owner address", undefined, types.string)
  .addParam("provider", "Staking Provider", undefined, types.string)
  .addParam("operator", "Staking Operator", undefined, types.string)
  .addOptionalParam("beneficiary", "Stake Beneficiary", undefined, types.string)
  .addOptionalParam("authorizer", "Stake Authorizer", undefined, types.string)
  .addOptionalParam("amount", "Stake amount", 1_000_000, types.int)
  .addOptionalParam(
    "authorization",
    "Authorization amount (default: minimumAuthorization)",
    undefined,
    types.int
  )
  .addFlag("skipBeacon", "Skip initialization for the Random Beacon contract")
  .setAction(async (args, hre) => {
    await initializeStake(hre, args)

    if (!args.skipBeacon) {
      await initializeBeacon(hre, args)
    }
  })

task(TASK_MINT, "Mints T tokens")
  .addParam("owner", "Stake Owner address", undefined, types.string)
  .addOptionalParam("amount", "Stake amount", 1_000_000, types.int)
  .setAction(async (args, hre) => {
    await mint(hre, args)
  })

task(TASK_STAKE, "Stakes T tokens")
  .addParam("owner", "Stake Owner address", undefined, types.string)
  .addParam("provider", "Staking Provider", undefined, types.string)
  .addOptionalParam("beneficiary", "Stake Beneficiary", undefined, types.string)
  .addOptionalParam("authorizer", "Stake Authorizer", undefined, types.string)
  .addOptionalParam("amount", "Stake amount", 1_000_000, types.int)
  .setAction(async (args, hre) => {
    await stake(hre, args)
  })

task(TASK_AUTHORIZE_BEACON, "Sets authorization for Beacon")
  .addParam("owner", "Stake Owner address", undefined, types.string)
  .addParam("provider", "Staking Provider", undefined, types.string)
  .addOptionalParam("authorizer", "Stake Authorizer", undefined, types.string)
  .addOptionalParam(
    "authorization",
    "Authorization amount (default: minimumAuthorization)",
    undefined,
    types.int
  )
  .setAction(async (args, hre) => {
    await authorize(hre, "RandomBeacon", args)
  })

task(
  TASK_REGISTER_BEACON,
  "Registers an operator for a staking provider in Beacon"
)
  .addParam("provider", "Staking Provider", undefined, types.string)
  .addParam("operator", "Staking Operator", undefined, types.string)
  .setAction(async (args, hre) => {
    await register(hre, "RandomBeacon", args)
  })

export async function initializeStake(
  hre: HardhatRuntimeEnvironment,
  args: {
    owner: string
    provider: string
    operator: string
    beneficiary: string
    authorizer: string
    amount: BigNumberish
    authorization: BigNumberish
  }
): Promise<void> {
  const tokensToMint = await calculateTokensNeededForStake(
    hre,
    args.provider,
    args.amount
  )

  if (!tokensToMint.isZero()) {
    await hre.run(TASK_MINT, { ...args, amount: tokensToMint.toNumber() })
  }

  await hre.run(TASK_STAKE, args)
}

async function initializeBeacon(
  hre: HardhatRuntimeEnvironment,
  args: {
    owner: string
    provider: string
    operator: string
    beneficiary: string
    authorizer: string
    amount: BigNumberish
    authorization: BigNumberish
  }
): Promise<void> {
  await hre.run(TASK_AUTHORIZE_BEACON, args)
  await hre.run(TASK_REGISTER_BEACON, args)
}
