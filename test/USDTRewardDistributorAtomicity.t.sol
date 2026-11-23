// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import {USDTRewardDistributor} from "../contracts/USDTRewardDistributor.sol";

contract ControlledMockUSDT {
    enum NestedAction {
        None,
        Distribute,
        Withdraw,
        ReplaceRecipients
    }

    mapping(address => uint256) public balanceOf;
    uint256 public transferCount;
    uint256 public failOnTransfer;
    NestedAction public nestedAction;
    bool public nestedSucceeded;
    bytes4 public nestedError;
    USDTRewardDistributor public distributor;
    address public nestedWithdrawalDestination;

    function mint(address account, uint256 amount) external {
        balanceOf[account] += amount;
    }

    function setFailOnTransfer(uint256 value) external {
        failOnTransfer = value;
    }

    function configureNestedAction(NestedAction action, address withdrawalDestination) external {
        nestedAction = action;
        nestedWithdrawalDestination = withdrawalDestination;
    }

    function setDistributor(USDTRewardDistributor value) external {
        distributor = value;
    }

    function callDistribute() external {
        distributor.distribute();
    }

    function callWithdraw(address destination) external {
        distributor.withdrawUSDT(destination);
    }

    function transfer(address recipient, uint256 amount) external returns (bool) {
        ++transferCount;
        if (nestedAction != NestedAction.None) {
            NestedAction action = nestedAction;
            nestedAction = NestedAction.None;
            _attemptNestedCall(action);
        }
        if (transferCount == failOnTransfer) {
            return false;
        }

        balanceOf[msg.sender] -= amount;
        balanceOf[recipient] += amount;
        return true;
    }

    function _attemptNestedCall(NestedAction action) private {
        if (action == NestedAction.Distribute) {
            try distributor.distribute() {
                nestedSucceeded = true;
            } catch (bytes memory reason) {
                nestedError = _selector(reason);
            }
        } else if (action == NestedAction.Withdraw) {
            try distributor.withdrawUSDT(nestedWithdrawalDestination) {
                nestedSucceeded = true;
            } catch (bytes memory reason) {
                nestedError = _selector(reason);
            }
        } else {
            USDTRewardDistributor.Recipient[] memory replacement = new USDTRewardDistributor.Recipient[](2);
            replacement[0] = USDTRewardDistributor.Recipient({account: address(0x11), shareBps: 9000});
            replacement[1] = USDTRewardDistributor.Recipient({account: address(0x22), shareBps: 1000});
            try distributor.replaceRecipients(replacement) {
                nestedSucceeded = true;
            } catch (bytes memory reason) {
                nestedError = _selector(reason);
            }
        }
    }

    function _selector(bytes memory reason) private pure returns (bytes4 selector) {
        if (reason.length < 4) {
            return bytes4(0);
        }
        assembly ("memory-safe") {
            selector := mload(add(reason, 32))
        }
    }
}

contract USDTRewardDistributorAtomicityTest {
    address private constant FEE_WALLET = address(0xFEE);
    address private constant RECIPIENT_ONE = address(0x11);
    address private constant RECIPIENT_TWO = address(0x22);
    address private constant DESTINATION = address(0xD357);
    address private constant NESTED_DESTINATION = address(0xD358);

    function testParticipantFailureRollsBackEarlierPayouts() external {
        (ControlledMockUSDT token, USDTRewardDistributor distributor) = _deploy(address(this), 0);
        token.mint(address(distributor), 100);
        token.setFailOnTransfer(2);

        _require(_distributeReverts(distributor), "participant failure accepted");
        _require(token.balanceOf(address(distributor)) == 100, "distributor balance changed");
        _require(token.balanceOf(RECIPIENT_ONE) == 0, "first recipient retained payout");
        _require(token.balanceOf(RECIPIENT_TWO) == 0, "second recipient paid");
        _require(token.balanceOf(FEE_WALLET) == 0, "fee paid");
        _require(token.transferCount() == 0, "token call state retained");
    }

    function testFeeFailureRollsBackParticipantPayouts() external {
        (ControlledMockUSDT token, USDTRewardDistributor distributor) = _deploy(address(this), 500);
        token.mint(address(distributor), 100);
        token.setFailOnTransfer(3);

        _require(_distributeReverts(distributor), "fee failure accepted");
        _require(token.balanceOf(address(distributor)) == 100, "distributor balance changed");
        _require(token.balanceOf(RECIPIENT_ONE) == 0, "first recipient retained payout");
        _require(token.balanceOf(RECIPIENT_TWO) == 0, "second recipient retained payout");
        _require(token.balanceOf(FEE_WALLET) == 0, "fee paid");
        _require(token.transferCount() == 0, "token call state retained");
    }

    function testWithdrawalFailureRollsBackBalance() external {
        (ControlledMockUSDT token, USDTRewardDistributor distributor) = _deploy(address(this), 0);
        token.mint(address(distributor), 100);
        token.setFailOnTransfer(1);

        _require(_withdrawReverts(distributor, DESTINATION), "withdrawal failure accepted");
        _require(token.balanceOf(address(distributor)) == 100, "distributor balance changed");
        _require(token.balanceOf(DESTINATION) == 0, "destination paid");
        _require(token.transferCount() == 0, "token call state retained");
    }

    function testNestedDistributionCannotSpendTwice() external {
        (ControlledMockUSDT token, USDTRewardDistributor distributor) = _deployWithTokenOwner();
        token.setDistributor(distributor);
        token.configureNestedAction(ControlledMockUSDT.NestedAction.Distribute, address(0));
        token.mint(address(distributor), 100);

        token.callDistribute();

        _require(!token.nestedSucceeded(), "nested distribution succeeded");
        _require(token.nestedError() == USDTRewardDistributor.ReentrantCall.selector, "wrong nested error");
        _require(token.balanceOf(RECIPIENT_ONE) == 50, "first payout changed");
        _require(token.balanceOf(RECIPIENT_TWO) == 50, "second payout changed");
        _require(token.balanceOf(address(distributor)) == 0, "distributor balance remains");
    }

    function testNestedWithdrawalCannotSpendTwice() external {
        (ControlledMockUSDT token, USDTRewardDistributor distributor) = _deployWithTokenOwner();
        token.setDistributor(distributor);
        token.configureNestedAction(ControlledMockUSDT.NestedAction.Withdraw, NESTED_DESTINATION);
        token.mint(address(distributor), 100);

        token.callWithdraw(DESTINATION);

        _require(!token.nestedSucceeded(), "nested withdrawal succeeded");
        _require(token.nestedError() == USDTRewardDistributor.ReentrantCall.selector, "wrong nested error");
        _require(token.balanceOf(DESTINATION) == 100, "withdrawal destination changed");
        _require(token.balanceOf(NESTED_DESTINATION) == 0, "nested withdrawal paid");
        _require(token.balanceOf(address(distributor)) == 0, "distributor balance remains");
    }

    function testNestedRecipientReplacementCannotAlterAllocation() external {
        (ControlledMockUSDT token, USDTRewardDistributor distributor) = _deployWithTokenOwner();
        token.setDistributor(distributor);
        token.configureNestedAction(ControlledMockUSDT.NestedAction.ReplaceRecipients, address(0));
        token.mint(address(distributor), 100);

        token.callDistribute();

        _require(!token.nestedSucceeded(), "nested replacement succeeded");
        _require(token.nestedError() == USDTRewardDistributor.ReentrantCall.selector, "wrong nested error");
        _require(token.balanceOf(RECIPIENT_ONE) == 50, "first payout changed");
        _require(token.balanceOf(RECIPIENT_TWO) == 50, "second payout changed");
        USDTRewardDistributor.Recipient[] memory recipients = distributor.getRecipients();
        _require(recipients[0].shareBps == 5000, "first allocation changed");
        _require(recipients[1].shareBps == 5000, "second allocation changed");
    }

    function _deploy(address owner, uint256 feeBps)
        private
        returns (ControlledMockUSDT token, USDTRewardDistributor distributor)
    {
        token = new ControlledMockUSDT();
        distributor = new USDTRewardDistributor(address(token), feeBps, owner, FEE_WALLET, _recipients());
    }

    function _deployWithTokenOwner() private returns (ControlledMockUSDT token, USDTRewardDistributor distributor) {
        token = new ControlledMockUSDT();
        distributor = new USDTRewardDistributor(address(token), 0, address(token), FEE_WALLET, _recipients());
    }

    function _distributeReverts(USDTRewardDistributor distributor) private returns (bool) {
        try distributor.distribute() {
            return false;
        } catch {
            return true;
        }
    }

    function _withdrawReverts(USDTRewardDistributor distributor, address destination) private returns (bool) {
        try distributor.withdrawUSDT(destination) {
            return false;
        } catch {
            return true;
        }
    }

    function _recipients() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](2);
        recipients[0] = USDTRewardDistributor.Recipient({account: RECIPIENT_ONE, shareBps: 5000});
        recipients[1] = USDTRewardDistributor.Recipient({account: RECIPIENT_TWO, shareBps: 5000});
    }

    function _require(bool condition, string memory message) private pure {
        require(condition, message);
    }
}
