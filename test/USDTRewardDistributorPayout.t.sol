// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import {USDTRewardDistributor} from "../contracts/USDTRewardDistributor.sol";

interface PayoutVm {
    function prank(address caller) external;
}

contract NoReturnMockUSDT {
    mapping(address => uint256) public balanceOf;
    mapping(address => uint256) public transferCalls;

    function mint(address account, uint256 amount) external {
        balanceOf[account] += amount;
    }

    function transfer(address recipient, uint256 amount) external {
        balanceOf[msg.sender] -= amount;
        balanceOf[recipient] += amount;
        ++transferCalls[recipient];
    }
}

contract BoolMockUSDT {
    mapping(address => uint256) public balanceOf;
    mapping(address => uint256) public transferCalls;
    bool public transferSucceeds = true;
    uint256 public transferCallCount;
    uint256 public transferReturnValue = 1;

    function mint(address account, uint256 amount) external {
        balanceOf[account] += amount;
    }

    function setTransferSucceeds(bool value) external {
        transferSucceeds = value;
    }

    function setTransferReturnValue(uint256 value) external {
        transferReturnValue = value;
    }

    function transfer(address recipient, uint256 amount) external returns (bool) {
        ++transferCallCount;
        if (!transferSucceeds) {
            return false;
        }

        balanceOf[msg.sender] -= amount;
        balanceOf[recipient] += amount;
        ++transferCalls[recipient];
        if (transferReturnValue != 1) {
            uint256 returnValue = transferReturnValue;
            assembly ("memory-safe") {
                mstore(0, returnValue)
                return(0, 32)
            }
        }
        return true;
    }
}

contract USDTRewardDistributorPayoutTest {
    PayoutVm private constant VM = PayoutVm(address(uint160(uint256(keccak256("hevm cheat code")))));

    address private constant OWNER = address(0xA11CE);
    address private constant FEE_WALLET = address(0xFEE);
    address private constant RECIPIENT_ONE = address(0x11);
    address private constant RECIPIENT_TWO = address(0x22);
    address private constant RECIPIENT_THREE = address(0x33);

    function testDistributeUsesExactAllocationWithBooleanReturningToken() external {
        BoolMockUSDT token = new BoolMockUSDT();
        USDTRewardDistributor distributor = _deploy(address(token), 500, FEE_WALLET, _threeRecipients());
        token.mint(address(distributor), 100_000_000);

        VM.prank(OWNER);
        distributor.distribute();

        _require(token.balanceOf(RECIPIENT_ONE) == 9_500_000, "first payout");
        _require(token.balanceOf(RECIPIENT_TWO) == 19_000_000, "second payout");
        _require(token.balanceOf(RECIPIENT_THREE) == 66_500_000, "third payout");
        _require(token.balanceOf(FEE_WALLET) == 5_000_000, "fee payout");
        _require(token.balanceOf(address(distributor)) == 0, "distributor balance");
    }

    function testDistributeSupportsTokenWithoutBooleanReturnValue() external {
        NoReturnMockUSDT token = new NoReturnMockUSDT();
        USDTRewardDistributor distributor = _deploy(address(token), 500, FEE_WALLET, _threeRecipients());
        token.mint(address(distributor), 100_000_000);

        VM.prank(OWNER);
        distributor.distribute();

        _require(token.balanceOf(RECIPIENT_ONE) == 9_500_000, "first payout");
        _require(token.balanceOf(RECIPIENT_TWO) == 19_000_000, "second payout");
        _require(token.balanceOf(RECIPIENT_THREE) == 66_500_000, "third payout");
        _require(token.balanceOf(FEE_WALLET) == 5_000_000, "fee payout");
        _require(token.balanceOf(address(distributor)) == 0, "distributor balance");
    }

    function testDistributeSendsFractionalResidueToFeeWallet() external {
        BoolMockUSDT token = new BoolMockUSDT();
        USDTRewardDistributor distributor = _deploy(address(token), 500, FEE_WALLET, _threeRecipients());
        token.mint(address(distributor), 100_000_001);

        VM.prank(OWNER);
        distributor.distribute();

        _require(token.balanceOf(RECIPIENT_ONE) == 9_500_000, "first payout");
        _require(token.balanceOf(RECIPIENT_TWO) == 19_000_000, "second payout");
        _require(token.balanceOf(RECIPIENT_THREE) == 66_500_000, "third payout");
        _require(token.balanceOf(FEE_WALLET) == 5_000_001, "fee residue");
        _require(token.balanceOf(address(distributor)) == 0, "distributor balance");
    }

    function testDistributeSkipsZeroPayoutsForTinyBalance() external {
        NoReturnMockUSDT token = new NoReturnMockUSDT();
        USDTRewardDistributor distributor = _deploy(address(token), 500, FEE_WALLET, _threeRecipients());
        token.mint(address(distributor), 7);

        VM.prank(OWNER);
        distributor.distribute();

        _require(token.balanceOf(RECIPIENT_ONE) == 0, "zero payout recipient paid");
        _require(token.transferCalls(RECIPIENT_ONE) == 0, "zero payout transfer called");
        _require(token.balanceOf(RECIPIENT_TWO) == 1, "second payout");
        _require(token.balanceOf(RECIPIENT_THREE) == 4, "third payout");
        _require(token.balanceOf(FEE_WALLET) == 2, "fee payout");
        _require(token.balanceOf(address(distributor)) == 0, "distributor balance");
    }

    function testDistributeAssignsZeroFeeResidueToFeeWallet() external {
        BoolMockUSDT token = new BoolMockUSDT();
        USDTRewardDistributor distributor = _deploy(address(token), 0, FEE_WALLET, _equalRecipients());
        token.mint(address(distributor), 1);

        VM.prank(OWNER);
        distributor.distribute();

        _require(token.balanceOf(RECIPIENT_ONE) == 0, "first recipient paid");
        _require(token.balanceOf(RECIPIENT_TWO) == 0, "second recipient paid");
        _require(token.balanceOf(FEE_WALLET) == 1, "fee residue");
        _require(token.balanceOf(address(distributor)) == 0, "distributor balance");
    }

    function testDistributeSkipsZeroFeeWalletPayoutWhenParticipantsConsumeBalance() external {
        BoolMockUSDT token = new BoolMockUSDT();
        USDTRewardDistributor distributor = _deploy(address(token), 0, FEE_WALLET, _equalRecipients());
        token.mint(address(distributor), 100);

        VM.prank(OWNER);
        distributor.distribute();

        _require(token.balanceOf(RECIPIENT_ONE) == 50, "first payout");
        _require(token.balanceOf(RECIPIENT_TWO) == 50, "second payout");
        _require(token.balanceOf(FEE_WALLET) == 0, "fee wallet paid");
        _require(token.transferCalls(FEE_WALLET) == 0, "zero fee wallet transfer called");
    }

    function testDistributeCombinesRecipientAndFeeWalletPayouts() external {
        BoolMockUSDT token = new BoolMockUSDT();
        USDTRewardDistributor distributor = _deploy(address(token), 500, FEE_WALLET, _feeWalletRecipient());
        token.mint(address(distributor), 100_000_000);

        VM.prank(OWNER);
        distributor.distribute();

        _require(token.balanceOf(RECIPIENT_ONE) == 47_500_000, "participant payout");
        _require(token.balanceOf(FEE_WALLET) == 52_500_000, "combined fee wallet payout");
        _require(token.balanceOf(address(distributor)) == 0, "distributor balance");
    }

    function testDistributeRejectsEmptyBalanceWithIdentifiableError() external {
        BoolMockUSDT token = new BoolMockUSDT();
        USDTRewardDistributor distributor = _deploy(address(token), 500, FEE_WALLET, _threeRecipients());

        VM.prank(OWNER);
        try distributor.distribute() {
            revert("empty balance accepted");
        } catch (bytes memory reason) {
            _require(_selector(reason) == USDTRewardDistributor.EmptyBalance.selector, "wrong empty balance error");
        }

        _require(token.transferCallCount() == 0, "token transfer called");
    }

    function testDistributeRejectsNonOwnerWithoutMovingFunds() external {
        BoolMockUSDT token = new BoolMockUSDT();
        USDTRewardDistributor distributor = _deploy(address(token), 500, FEE_WALLET, _threeRecipients());
        token.mint(address(distributor), 100_000_000);

        VM.prank(address(0xBAD));
        _require(_distributeReverts(distributor), "non-owner distributed");

        _require(token.balanceOf(address(distributor)) == 100_000_000, "balance changed");
        _require(token.balanceOf(RECIPIENT_ONE) == 0, "recipient paid");
        _require(token.balanceOf(FEE_WALLET) == 0, "fee paid");
    }

    function testDistributeRejectsTokenReportedFailure() external {
        BoolMockUSDT token = new BoolMockUSDT();
        USDTRewardDistributor distributor = _deploy(address(token), 500, FEE_WALLET, _threeRecipients());
        token.mint(address(distributor), 100_000_000);
        token.setTransferSucceeds(false);

        VM.prank(OWNER);
        _require(_distributeReverts(distributor), "token failure accepted");
        _require(token.balanceOf(address(distributor)) == 100_000_000, "balance changed");
    }

    function testDistributeRejectsMalformedTokenReturnAtomically() external {
        BoolMockUSDT token = new BoolMockUSDT();
        USDTRewardDistributor distributor = _deploy(address(token), 500, FEE_WALLET, _threeRecipients());
        token.mint(address(distributor), 100_000_000);
        token.setTransferReturnValue(2);

        VM.prank(OWNER);
        _require(_distributeReverts(distributor), "malformed return accepted");

        _require(token.balanceOf(address(distributor)) == 100_000_000, "distributor balance changed");
        _require(token.balanceOf(RECIPIENT_ONE) == 0, "first recipient paid");
        _require(token.balanceOf(RECIPIENT_TWO) == 0, "second recipient paid");
        _require(token.balanceOf(RECIPIENT_THREE) == 0, "third recipient paid");
        _require(token.balanceOf(FEE_WALLET) == 0, "fee paid");
        _require(token.transferCallCount() == 0, "token call state retained");
    }

    function _deploy(
        address token,
        uint256 feeBps,
        address feeWallet,
        USDTRewardDistributor.Recipient[] memory recipients
    ) private returns (USDTRewardDistributor) {
        return new USDTRewardDistributor(token, feeBps, OWNER, feeWallet, recipients);
    }

    function _distributeReverts(USDTRewardDistributor distributor) private returns (bool) {
        try distributor.distribute() {
            return false;
        } catch {
            return true;
        }
    }

    function _threeRecipients() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](3);
        recipients[0] = USDTRewardDistributor.Recipient({account: RECIPIENT_ONE, shareBps: 1000});
        recipients[1] = USDTRewardDistributor.Recipient({account: RECIPIENT_TWO, shareBps: 2000});
        recipients[2] = USDTRewardDistributor.Recipient({account: RECIPIENT_THREE, shareBps: 7000});
    }

    function _equalRecipients() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](2);
        recipients[0] = USDTRewardDistributor.Recipient({account: RECIPIENT_ONE, shareBps: 5000});
        recipients[1] = USDTRewardDistributor.Recipient({account: RECIPIENT_TWO, shareBps: 5000});
    }

    function _feeWalletRecipient() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](2);
        recipients[0] = USDTRewardDistributor.Recipient({account: RECIPIENT_ONE, shareBps: 5000});
        recipients[1] = USDTRewardDistributor.Recipient({account: FEE_WALLET, shareBps: 5000});
    }

    function _selector(bytes memory reason) private pure returns (bytes4 selector) {
        _require(reason.length >= 4, "missing error selector");
        assembly ("memory-safe") {
            selector := mload(add(reason, 32))
        }
    }

    function _require(bool condition, string memory message) private pure {
        require(condition, message);
    }
}
