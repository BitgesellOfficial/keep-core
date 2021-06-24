import React from "react"
import { useDispatch } from "react-redux"
import PageWrapper from "../../components/PageWrapper"
import {
  CheckListBanner,
  HowDoesItWorkBanner,
  DepositForm,
  InitiateDepositModal,
} from "../../components/coverage-pools"
import TokenAmount from "../../components/TokenAmount"
import MetricsTile from "../../components/MetricsTile"
import { APY } from "../../components/liquidity"
import { KEEP } from "../../utils/token.utils"
import { useModal } from "../../hooks/useModal"
import { depositAssetPool } from "../../actions/web3"
import WithdrawAmountForm from "../../components/WithdrawAmountForm"

const CoveragePoolPage = ({ title, withNewLabel }) => {
  const shareOfPool = "0"
  const rewards = "0"

  const { openConfirmationModal } = useModal()
  const dispatch = useDispatch()

  const onSubmitDepositForm = async (values, awaitingPromise) => {
    const { tokenAmount } = values
    const amount = KEEP.fromTokenUnit(tokenAmount)
    await openConfirmationModal(
      {
        modalOptions: { title: "Initiate Deposit" },
        submitBtnText: "deposit",
        amount,
      },
      InitiateDepositModal
    )
    dispatch(depositAssetPool(amount, awaitingPromise))
  }

  const onMaxAmountClick = () => {}

  const onCancel = () => {}

  const onSubmit = () => {}

  return (
    <PageWrapper title={title} newPage={withNewLabel}>
      <CheckListBanner />

      <section className="tile coverage-pool__overview">
        <section className="coverage-pool__overview__tvl">
          <h2 className="h2--alt text-grey-70 mb-1">Total Value Locked</h2>
          <TokenAmount
            amount="900000000000000000000000000"
            amountClassName="h1 text-mint-100"
            symbolClassName="h2 text-mint-100"
            withIcon
          />
        </section>
        <section className="coverage-pool__overview__apy">
          <h3 className="text-grey-70 mb-1">Pool APY</h3>
          <section className="apy__values">
            <MetricsTile className="bg-mint-10">
              <APY apy="0.15" className="text-mint-100" />
              <h5 className="text-grey-60">weekly</h5>
            </MetricsTile>
            <MetricsTile className="bg-mint-10">
              <APY apy="0.50" className="text-mint-100 " />
              <h5 className="text-grey-60">monthly</h5>
            </MetricsTile>
            <MetricsTile className="bg-mint-10">
              <APY apy="1.40" className="text-mint-100" />
              <h5 className="text-grey-60">annual</h5>
            </MetricsTile>
          </section>
        </section>
      </section>

      <section className="coverage-pool__deposit-wrapper">
        <section className="tile coverage-pool__deposit-form">
          <h3>Deposit</h3>
          <DepositForm onSubmit={onSubmitDepositForm} />
        </section>

        <section className="tile coverage-pool__share-of-pool">
          <h4 className="text-grey-70">Your Share of Pool</h4>
        </section>

        <section className="tile coverage-pool__rewards">
          <h4 className="text-grey-70">Your Rewards</h4>
        </section>

        <section className="tile coverage-pool__withdraw-wrapper">
          <h3>Available to withdraw</h3>
          <TokenAmount
            wrapperClassName={"coverage-pool__token-amount"}
            amount={"100000000000000000000000"}
            withIcon
          />
          <WithdrawAmountForm
            onCancel={onCancel}
            submitBtnText="add keep"
            availableAmount={"10000000000000000"}
            currentAmount={"10000000000000000"}
            onBtnClick={onSubmit}
          />
        </section>
      </section>

      <section className={"coverage-pool__pending-withdrawal"}></section>
    </PageWrapper>
  )
}

export default CoveragePoolPage
