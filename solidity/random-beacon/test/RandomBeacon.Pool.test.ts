/* eslint-disable @typescript-eslint/no-unused-expressions */

import { ethers, waffle, helpers } from "hardhat"
import { expect } from "chai"
import { ContractTransaction } from "ethers"
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers"
import { randomBeaconDeployment, constants } from "./fixtures"
import type { RandomBeaconStub, SortitionPool, StakingStub } from "../typechain"

const { time } = helpers
const { increaseTime } = time

const fixture = async () => randomBeaconDeployment()

describe("RandomBeacon - Pool", () => {
  let operator: SignerWithAddress
  let randomBeacon: RandomBeaconStub
  let sortitionPool: SortitionPool
  let stakingStub: StakingStub

  // prettier-ignore
  before(async () => {
    [operator] = await ethers.getSigners()
  })

  beforeEach("load test fixture", async () => {
    const contracts = await waffle.loadFixture(fixture)

    randomBeacon = contracts.randomBeacon as RandomBeaconStub
    sortitionPool = contracts.sortitionPool as SortitionPool
    stakingStub = contracts.stakingStub as StakingStub
  })

  describe("registerOperator", () => {
    beforeEach(async () => {
      await stakingStub.setStake(operator.address, constants.minimumStake)
    })

    context("when the operator is not registered yet", () => {
      context("when there is no active punishment for given operator", () => {
        beforeEach(async () => {
          await randomBeacon.connect(operator).registerOperator()
        })

        it("should deposit gas", async () => {
          expect(await randomBeacon.hasGasDeposit(operator.address)).to.be.true
        })

        it("should register the operator", async () => {
          expect(await sortitionPool.isOperatorInPool(operator.address)).to.be
            .true
        })
      })

      context("when punishment for given operator already elapsed", () => {
        let operatorID: number

        beforeEach(async () => {
          await randomBeacon.connect(operator).registerOperator()
          operatorID = await sortitionPool.getOperatorID(operator.address)

          const punishmentDuration = 1209600 // 2 weeks
          await randomBeacon.publicPunishOperators(
            [operatorID],
            punishmentDuration
          )

          await increaseTime(punishmentDuration)

          await randomBeacon.connect(operator).registerOperator()
        })

        it("should deposit gas", async () => {
          expect(await randomBeacon.hasGasDeposit(operator.address)).to.be.true
        })

        it("should register the operator", async () => {
          expect(await sortitionPool.isOperatorInPool(operator.address)).to.be
            .true
        })

        it("should remove operator from punished operators map", async () => {
          expect(
            await randomBeacon.punishedOperators(operator.address)
          ).to.be.equal(0)
        })
      })

      context("when there is an active punishment for given operator", () => {
        let operatorID: number

        beforeEach(async () => {
          await randomBeacon.connect(operator).registerOperator()
          operatorID = await sortitionPool.getOperatorID(operator.address)

          const punishmentDuration = 1209600 // 2 weeks
          await randomBeacon.publicPunishOperators(
            [operatorID],
            punishmentDuration
          )

          await increaseTime(punishmentDuration - 1)
        })

        it("should revert", async () => {
          await expect(
            randomBeacon.connect(operator).registerOperator()
          ).to.be.revertedWith("Operator has an active punishment")
        })
      })
    })

    context("when the operator is already registered", () => {
      beforeEach(async () => {
        await randomBeacon.connect(operator).registerOperator()
      })

      it("should revert", async () => {
        await expect(
          randomBeacon.connect(operator).registerOperator()
        ).to.be.revertedWith("Operator is already registered")
      })
    })
  })

  describe("updateOperatorStatus", () => {
    let operatorID: number

    beforeEach(async () => {
      // Operator is registered and gas deposit is made.
      await stakingStub.setStake(operator.address, constants.minimumStake)
      await randomBeacon.connect(operator).registerOperator()
      operatorID = await sortitionPool.getOperatorID(operator.address)

      // Simulate the operator became ineligible.
      await stakingStub.setStake(operator.address, 0)

      await randomBeacon.connect(operator).updateOperatorStatus()
    })

    it("should remove operator from the pool", async () => {
      expect(await sortitionPool.isOperatorInPool(operator.address)).to.be.false
    })

    it("should release the gas deposit if operator was removed from pool during the update", async () => {
      expect(await randomBeacon.hasGasDeposit(operator.address)).to.be.false
    })
  })

  describe("isOperatorEligible", () => {
    context("when the operator is eligible to join the sortition pool", () => {
      beforeEach(async () => {
        await stakingStub.setStake(operator.address, constants.minimumStake)
      })

      it("should return true", async () => {
        await expect(await randomBeacon.isOperatorEligible(operator.address)).to
          .be.true
      })
    })

    context(
      "when the operator is not eligible to join the sortition pool",
      () => {
        beforeEach(async () => {
          await stakingStub.setStake(
            operator.address,
            constants.minimumStake.sub(1)
          )
        })

        it("should return false", async () => {
          await expect(await randomBeacon.isOperatorEligible(operator.address))
            .to.be.false
        })
      }
    )
  })
})
