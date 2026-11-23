// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import {USDTRewardDistributor} from "../contracts/USDTRewardDistributor.sol";

interface OwnershipVm {
    function prank(address caller) external;
}

contract OwnershipMockUSDT {}

contract USDTRewardDistributorOwnershipTest {
    OwnershipVm private constant VM = OwnershipVm(address(uint160(uint256(keccak256("hevm cheat code")))));

    address private constant NEW_OWNER = address(0xA11CE);
    address private constant NEXT_OWNER = address(0xB0B);
    address private constant INITIAL_FEE_WALLET = address(0xFEE);
    address private constant NEW_FEE_WALLET = address(0xCAFE);

    OwnershipMockUSDT private token = new OwnershipMockUSDT();
    USDTRewardDistributor private distributor;

    function setUp() external {
        distributor = _deploy();
    }

    function testOwnershipTransferIsImmediateAndPreservesConfiguration() external {
        distributor.transferOwnership(NEW_OWNER);

        _require(distributor.owner() == NEW_OWNER, "new owner not installed");
        _require(distributor.feeWallet() == INITIAL_FEE_WALLET, "fee wallet changed");
        _require(distributor.feeBps() == 500, "fee rate changed");
        _requireInitialRecipients();
    }

    function testNewOwnerGetsAllOwnerPowersAndPreviousOwnerLosesAccess() external {
        distributor.transferOwnership(NEW_OWNER);

        _require(_replacementReverts(_oneRecipient()), "previous owner replaced recipients");
        _require(_feeWalletUpdateReverts(NEW_FEE_WALLET), "previous owner updated fee wallet");
        _require(_ownershipTransferReverts(NEXT_OWNER), "previous owner transferred ownership");

        VM.prank(NEW_OWNER);
        distributor.replaceRecipients(_oneRecipient());
        USDTRewardDistributor.Recipient[] memory recipients = distributor.getRecipients();
        _require(recipients.length == 1, "new owner did not replace recipients");
        _requireRecipient(recipients, 0, address(0x11), 10_000);

        VM.prank(NEW_OWNER);
        distributor.setFeeWallet(NEW_FEE_WALLET);
        _require(distributor.feeWallet() == NEW_FEE_WALLET, "new owner did not update fee wallet");

        VM.prank(NEW_OWNER);
        distributor.transferOwnership(NEXT_OWNER);
        _require(distributor.owner() == NEXT_OWNER, "new owner did not transfer ownership");
    }

    function testOwnershipTransferRejectsZeroAndDistributorAddresses() external {
        _require(_ownershipTransferReverts(address(0)), "zero owner accepted");
        _require(_ownershipTransferReverts(address(distributor)), "distributor owner accepted");
        _requireInitialConfiguration();
    }

    function testOwnerCanUpdateFeeWalletWithoutChangingOtherConfiguration() external {
        distributor.setFeeWallet(NEW_FEE_WALLET);

        _require(distributor.owner() == address(this), "owner changed");
        _require(distributor.feeWallet() == NEW_FEE_WALLET, "fee wallet not updated");
        _require(distributor.feeBps() == 500, "fee rate changed");
        _requireInitialRecipients();
    }

    function testFeeWalletUpdateRejectsZeroAndDistributorAddresses() external {
        _require(_feeWalletUpdateReverts(address(0)), "zero fee wallet accepted");
        _requireInitialConfiguration();

        _require(_feeWalletUpdateReverts(address(distributor)), "distributor fee wallet accepted");
        _requireInitialConfiguration();
    }

    function testUnauthorizedConfigurationCallsPreserveConfiguration() external {
        VM.prank(address(0xBAD));
        _require(_replacementReverts(_oneRecipient()), "unauthorized replacement accepted");
        _requireInitialConfiguration();

        VM.prank(address(0xBAD));
        _require(_feeWalletUpdateReverts(NEW_FEE_WALLET), "unauthorized fee wallet update accepted");
        _requireInitialConfiguration();

        VM.prank(address(0xBAD));
        _require(_ownershipTransferReverts(NEW_OWNER), "unauthorized ownership transfer accepted");
        _requireInitialConfiguration();
    }

    function _deploy() private returns (USDTRewardDistributor) {
        return new USDTRewardDistributor(address(token), 500, address(this), INITIAL_FEE_WALLET, _initialRecipients());
    }

    function _replacementReverts(USDTRewardDistributor.Recipient[] memory replacement) private returns (bool) {
        try distributor.replaceRecipients(replacement) {
            return false;
        } catch {
            return true;
        }
    }

    function _feeWalletUpdateReverts(address newFeeWallet) private returns (bool) {
        try distributor.setFeeWallet(newFeeWallet) {
            return false;
        } catch {
            return true;
        }
    }

    function _ownershipTransferReverts(address newOwner) private returns (bool) {
        try distributor.transferOwnership(newOwner) {
            return false;
        } catch {
            return true;
        }
    }

    function _requireInitialConfiguration() private view {
        _require(distributor.usdt() == address(token), "USDT changed");
        _require(distributor.feeBps() == 500, "fee rate changed");
        _require(distributor.owner() == address(this), "owner changed");
        _require(distributor.feeWallet() == INITIAL_FEE_WALLET, "fee wallet changed");
        _requireInitialRecipients();
    }

    function _requireInitialRecipients() private view {
        USDTRewardDistributor.Recipient[] memory recipients = distributor.getRecipients();
        _require(recipients.length == 2, "recipient count changed");
        _requireRecipient(recipients, 0, address(3), 5000);
        _requireRecipient(recipients, 1, address(4), 5000);
    }

    function _requireRecipient(
        USDTRewardDistributor.Recipient[] memory recipients,
        uint256 index,
        address account,
        uint16 shareBps
    ) private pure {
        _require(recipients[index].account == account, "unexpected recipient account");
        _require(recipients[index].shareBps == shareBps, "unexpected recipient share");
    }

    function _initialRecipients() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](2);
        recipients[0] = USDTRewardDistributor.Recipient({account: address(3), shareBps: 5000});
        recipients[1] = USDTRewardDistributor.Recipient({account: address(4), shareBps: 5000});
    }

    function _oneRecipient() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](1);
        recipients[0] = USDTRewardDistributor.Recipient({account: address(0x11), shareBps: 10_000});
    }

    function _require(bool condition, string memory message) private pure {
        require(condition, message);
    }
}
