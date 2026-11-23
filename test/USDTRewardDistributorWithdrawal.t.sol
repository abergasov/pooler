// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import {USDTRewardDistributor} from "../contracts/USDTRewardDistributor.sol";

interface WithdrawalVm {
    function prank(address caller) external;
}

contract WithdrawalMockUSDT {
    error TransferReverted();

    mapping(address => uint256) public balanceOf;
    mapping(address => uint256) public transferCalls;
    bool public transferSucceeds = true;
    bool public transferReverts;
    bool public transferMovesFunds = true;
    uint256 public transferReturnValue = 1;

    function mint(address account, uint256 amount) external {
        balanceOf[account] += amount;
    }

    function setTransferSucceeds(bool value) external {
        transferSucceeds = value;
    }

    function setTransferReverts(bool value) external {
        transferReverts = value;
    }

    function setTransferMovesFunds(bool value) external {
        transferMovesFunds = value;
    }

    function setTransferReturnValue(uint256 value) external {
        transferReturnValue = value;
    }

    function transfer(address recipient, uint256 amount) external returns (bool) {
        if (transferReverts) {
            revert TransferReverted();
        }
        if (!transferSucceeds) {
            return false;
        }
        if (!transferMovesFunds) {
            return true;
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

contract USDTRewardDistributorWithdrawalTest {
    WithdrawalVm private constant VM = WithdrawalVm(address(uint160(uint256(keccak256("hevm cheat code")))));

    address private constant OWNER = address(0xA11CE);
    address private constant FEE_WALLET = address(0xFEE);
    address private constant RECIPIENT_ONE = address(0x11);
    address private constant RECIPIENT_TWO = address(0x22);
    address private constant DESTINATION = address(0xD357);

    WithdrawalMockUSDT private token;
    USDTRewardDistributor private distributor;

    function setUp() external {
        token = new WithdrawalMockUSDT();
        distributor = new USDTRewardDistributor(address(token), 500, OWNER, FEE_WALLET, _recipients());
    }

    function testWithdrawTransfersFullBalanceWithoutPayoutsOrFees() external {
        token.mint(address(distributor), 100_000_000);

        VM.prank(OWNER);
        distributor.withdrawUSDT(DESTINATION);

        _require(token.balanceOf(DESTINATION) == 100_000_000, "destination balance");
        _require(token.balanceOf(address(distributor)) == 0, "distributor balance");
        _require(token.balanceOf(RECIPIENT_ONE) == 0, "first recipient paid");
        _require(token.balanceOf(RECIPIENT_TWO) == 0, "second recipient paid");
        _require(token.balanceOf(FEE_WALLET) == 0, "fee paid");
        _requireInitialConfiguration();
    }

    function testWithdrawRejectsEmptyBalanceWithoutTokenCall() external {
        VM.prank(OWNER);
        try distributor.withdrawUSDT(DESTINATION) {
            revert("empty balance accepted");
        } catch (bytes memory reason) {
            _require(_selector(reason) == USDTRewardDistributor.EmptyBalance.selector, "wrong empty balance error");
        }

        _require(token.transferCalls(DESTINATION) == 0, "token transfer called");
    }

    function testWithdrawRejectsNonOwnerWithoutMovingFunds() external {
        token.mint(address(distributor), 100_000_000);

        VM.prank(address(0xBAD));
        _require(_withdrawReverts(DESTINATION), "non-owner withdrew");

        _require(token.balanceOf(address(distributor)) == 100_000_000, "balance changed");
        _require(token.balanceOf(DESTINATION) == 0, "destination paid");
    }

    function testWithdrawRejectsZeroAndSelfDestinationsWithoutMovingFunds() external {
        token.mint(address(distributor), 100_000_000);

        VM.prank(OWNER);
        _require(
            _withdrawError(address(0)) == USDTRewardDistributor.InvalidWithdrawalDestination.selector,
            "wrong zero destination error"
        );
        VM.prank(OWNER);
        _require(
            _withdrawError(address(distributor)) == USDTRewardDistributor.InvalidWithdrawalDestination.selector,
            "wrong self destination error"
        );

        _require(token.balanceOf(address(distributor)) == 100_000_000, "balance changed");
        _require(token.transferCalls(address(0)) == 0, "zero destination transfer called");
        _require(token.transferCalls(address(distributor)) == 0, "self destination transfer called");
    }

    function testWithdrawRevertsWhenTokenReturnsFalse() external {
        token.mint(address(distributor), 100_000_000);
        token.setTransferSucceeds(false);

        VM.prank(OWNER);
        _require(
            _withdrawError(DESTINATION) == USDTRewardDistributor.TokenTransferFailed.selector,
            "wrong token failure error"
        );

        _require(token.balanceOf(address(distributor)) == 100_000_000, "balance changed");
        _require(token.balanceOf(DESTINATION) == 0, "destination paid");
    }

    function testWithdrawRevertsWhenTokenTransferReverts() external {
        token.mint(address(distributor), 100_000_000);
        token.setTransferReverts(true);

        VM.prank(OWNER);
        _require(
            _withdrawError(DESTINATION) == USDTRewardDistributor.TokenTransferFailed.selector,
            "wrong token revert error"
        );

        _require(token.balanceOf(address(distributor)) == 100_000_000, "balance changed");
        _require(token.balanceOf(DESTINATION) == 0, "destination paid");
    }

    function testWithdrawRevertsWhenTokenReturnsMalformedBoolean() external {
        token.mint(address(distributor), 100_000_000);
        token.setTransferReturnValue(2);

        VM.prank(OWNER);
        _require(_withdrawReverts(DESTINATION), "malformed return accepted");

        _require(token.balanceOf(address(distributor)) == 100_000_000, "distributor balance changed");
        _require(token.balanceOf(DESTINATION) == 0, "destination paid");
        _require(token.transferCalls(DESTINATION) == 0, "token call state retained");
    }

    function testWithdrawRevertsWhenTokenLeavesBalance() external {
        token.mint(address(distributor), 100_000_000);
        token.setTransferMovesFunds(false);

        VM.prank(OWNER);
        _require(
            _withdrawError(DESTINATION) == USDTRewardDistributor.RemainingBalance.selector,
            "wrong remaining balance error"
        );

        _require(token.balanceOf(address(distributor)) == 100_000_000, "balance changed");
        _require(token.balanceOf(DESTINATION) == 0, "destination paid");
    }

    function _withdrawReverts(address destination) private returns (bool) {
        try distributor.withdrawUSDT(destination) {
            return false;
        } catch {
            return true;
        }
    }

    function _withdrawError(address destination) private returns (bytes4) {
        try distributor.withdrawUSDT(destination) {
            revert("withdraw succeeded");
        } catch (bytes memory reason) {
            return _selector(reason);
        }
    }

    function _requireInitialConfiguration() private view {
        _require(distributor.owner() == OWNER, "owner changed");
        _require(distributor.feeWallet() == FEE_WALLET, "fee wallet changed");
        _require(distributor.feeBps() == 500, "fee changed");
        _require(distributor.usdt() == address(token), "token changed");

        USDTRewardDistributor.Recipient[] memory recipients = distributor.getRecipients();
        _require(recipients.length == 2, "recipient count changed");
        _require(recipients[0].account == RECIPIENT_ONE && recipients[0].shareBps == 4_000, "first recipient changed");
        _require(recipients[1].account == RECIPIENT_TWO && recipients[1].shareBps == 6_000, "second recipient changed");
    }

    function _recipients() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](2);
        recipients[0] = USDTRewardDistributor.Recipient({account: RECIPIENT_ONE, shareBps: 4_000});
        recipients[1] = USDTRewardDistributor.Recipient({account: RECIPIENT_TWO, shareBps: 6_000});
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
