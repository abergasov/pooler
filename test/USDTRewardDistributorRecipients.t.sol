// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import {USDTRewardDistributor} from "../contracts/USDTRewardDistributor.sol";

interface Vm {
    function prank(address caller) external;
}

contract MockUSDT {}

contract USDTRewardDistributorRecipientsTest {
    Vm private constant VM = Vm(address(uint160(uint256(keccak256("hevm cheat code")))));

    MockUSDT private token = new MockUSDT();
    USDTRewardDistributor private distributor;

    function setUp() external {
        distributor = _deploy();
    }

    function testOwnerCanReplaceEntireRecipientList() external {
        USDTRewardDistributor.Recipient[] memory replacement = new USDTRewardDistributor.Recipient[](3);
        replacement[0] = USDTRewardDistributor.Recipient({account: address(0x11), shareBps: 1000});
        replacement[1] = USDTRewardDistributor.Recipient({account: address(0x22), shareBps: 2000});
        replacement[2] = USDTRewardDistributor.Recipient({account: address(0x33), shareBps: 7000});

        distributor.replaceRecipients(replacement);

        USDTRewardDistributor.Recipient[] memory recipients = distributor.getRecipients();
        _require(recipients.length == 3, "replacement count");
        _requireRecipient(recipients, 0, address(0x11), 1000);
        _requireRecipient(recipients, 1, address(0x22), 2000);
        _requireRecipient(recipients, 2, address(0x33), 7000);
    }

    function testOwnerCanReplaceWithOneRecipient() external {
        USDTRewardDistributor.Recipient[] memory replacement = new USDTRewardDistributor.Recipient[](1);
        replacement[0] = USDTRewardDistributor.Recipient({account: address(0x11), shareBps: 10_000});

        distributor.replaceRecipients(replacement);

        USDTRewardDistributor.Recipient[] memory recipients = distributor.getRecipients();
        _require(recipients.length == 1, "one-recipient count");
        _requireRecipient(recipients, 0, address(0x11), 10_000);
    }

    function testReplacementAllowsOwnerAndFeeWalletAsRecipients() external {
        USDTRewardDistributor.Recipient[] memory replacement = new USDTRewardDistributor.Recipient[](2);
        replacement[0] = USDTRewardDistributor.Recipient({account: address(this), shareBps: 5000});
        replacement[1] = USDTRewardDistributor.Recipient({account: address(0xFEE), shareBps: 5000});

        distributor.replaceRecipients(replacement);

        USDTRewardDistributor.Recipient[] memory recipients = distributor.getRecipients();
        _require(recipients.length == 2, "owner and fee wallet count");
        _requireRecipient(recipients, 0, address(this), 5000);
        _requireRecipient(recipients, 1, address(0xFEE), 5000);
    }

    function testOwnerCanReplaceWithMaximumRecipients() external {
        USDTRewardDistributor.Recipient[] memory replacement = _maximumRecipients();

        distributor.replaceRecipients(replacement);

        USDTRewardDistributor.Recipient[] memory recipients = distributor.getRecipients();
        _require(recipients.length == 100, "maximum-recipient count");
        _requireRecipient(recipients, 0, address(1), 100);
        _requireRecipient(recipients, 99, address(100), 100);
    }

    function testUnauthorizedReplacementPreservesRecipients() external {
        USDTRewardDistributor.Recipient[] memory replacement = _oneRecipient();

        VM.prank(address(0xBAD));
        _require(_replacementReverts(replacement), "unauthorized replacement accepted");
        _requireInitialConfiguration();
    }

    function testInvalidReplacementsPreserveRecipients() external {
        USDTRewardDistributor.Recipient[] memory empty = new USDTRewardDistributor.Recipient[](0);
        _require(_replacementReverts(empty), "empty replacement accepted");
        _requireInitialConfiguration();

        USDTRewardDistributor.Recipient[] memory tooMany = new USDTRewardDistributor.Recipient[](101);
        for (uint256 i; i < 101; ++i) {
            tooMany[i] = USDTRewardDistributor.Recipient({account: address(uint160(i + 1)), shareBps: 100});
        }
        _require(_replacementReverts(tooMany), "101-recipient replacement accepted");
        _requireInitialConfiguration();

        USDTRewardDistributor.Recipient[] memory duplicate = _twoRecipients();
        duplicate[1].account = duplicate[0].account;
        _require(_replacementReverts(duplicate), "duplicate replacement accepted");
        _requireInitialConfiguration();

        USDTRewardDistributor.Recipient[] memory zeroRecipient = _twoRecipients();
        zeroRecipient[0].account = address(0);
        _require(_replacementReverts(zeroRecipient), "zero-recipient replacement accepted");
        _requireInitialConfiguration();

        USDTRewardDistributor.Recipient[] memory selfRecipient = _twoRecipients();
        selfRecipient[0].account = address(distributor);
        _require(_replacementReverts(selfRecipient), "self-recipient replacement accepted");
        _requireInitialConfiguration();

        USDTRewardDistributor.Recipient[] memory zeroShare = _twoRecipients();
        zeroShare[0].shareBps = 0;
        _require(_replacementReverts(zeroShare), "zero-share replacement accepted");
        _requireInitialConfiguration();

        USDTRewardDistributor.Recipient[] memory lowTotal = _twoRecipients();
        lowTotal[1].shareBps = 4999;
        _require(_replacementReverts(lowTotal), "low-total replacement accepted");
        _requireInitialConfiguration();

        USDTRewardDistributor.Recipient[] memory highTotal = _twoRecipients();
        highTotal[1].shareBps = 5001;
        _require(_replacementReverts(highTotal), "high-total replacement accepted");
        _requireInitialConfiguration();
    }

    function _deploy() private returns (USDTRewardDistributor) {
        return new USDTRewardDistributor(address(token), 0, address(this), address(0xFEE), _twoRecipients());
    }

    function _replacementReverts(USDTRewardDistributor.Recipient[] memory replacement) private returns (bool) {
        try distributor.replaceRecipients(replacement) {
            return false;
        } catch {
            return true;
        }
    }

    function _requireInitialConfiguration() private view {
        _require(distributor.usdt() == address(token), "USDT changed");
        _require(distributor.feeBps() == 0, "fee rate changed");
        _require(distributor.owner() == address(this), "owner changed");
        _require(distributor.feeWallet() == address(0xFEE), "fee wallet changed");
        USDTRewardDistributor.Recipient[] memory recipients = distributor.getRecipients();
        _require(recipients.length == 2, "initial recipient count changed");
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

    function _oneRecipient() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](1);
        recipients[0] = USDTRewardDistributor.Recipient({account: address(0x11), shareBps: 10_000});
    }

    function _twoRecipients() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](2);
        recipients[0] = USDTRewardDistributor.Recipient({account: address(3), shareBps: 5000});
        recipients[1] = USDTRewardDistributor.Recipient({account: address(4), shareBps: 5000});
    }

    function _maximumRecipients() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](100);
        for (uint256 i; i < 100; ++i) {
            recipients[i] = USDTRewardDistributor.Recipient({account: address(uint160(i + 1)), shareBps: 100});
        }
    }

    function _require(bool condition, string memory message) private pure {
        require(condition, message);
    }
}
