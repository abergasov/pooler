// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import {USDTRewardDistributor} from "../contracts/USDTRewardDistributor.sol";

interface ConfigurationVm {
    function prank(address caller) external;
}

contract ConfigurationMockUSDT {}

contract USDTRewardDistributorConfigurationTest {
    ConfigurationVm private constant VM = ConfigurationVm(address(uint160(uint256(keccak256("hevm cheat code")))));

    address private constant OWNER = address(0xA11CE);
    address private constant NEW_OWNER = address(0xB0B);
    address private constant INITIAL_FEE_WALLET = address(0xFEE);
    address private constant UPDATED_FEE_WALLET = address(0xCAFE);

    ConfigurationMockUSDT private token = new ConfigurationMockUSDT();
    USDTRewardDistributor private distributor;

    function setUp() external {
        distributor = new USDTRewardDistributor(address(token), 500, OWNER, INITIAL_FEE_WALLET, _initialRecipients());
    }

    function testUnprivilegedReaderCanReadInitialConfiguration() external view {
        _requireConfiguration(OWNER, INITIAL_FEE_WALLET, _initialRecipients());
    }

    function testUnprivilegedReaderCanReadUpdatedConfiguration() external {
        USDTRewardDistributor.Recipient[] memory updatedRecipients = _updatedRecipients();

        VM.prank(OWNER);
        distributor.replaceRecipients(updatedRecipients);
        VM.prank(OWNER);
        distributor.transferOwnership(NEW_OWNER);
        VM.prank(NEW_OWNER);
        distributor.setFeeWallet(UPDATED_FEE_WALLET);

        _requireConfiguration(NEW_OWNER, UPDATED_FEE_WALLET, updatedRecipients);
    }

    function _requireConfiguration(
        address expectedOwner,
        address expectedFeeWallet,
        USDTRewardDistributor.Recipient[] memory expectedRecipients
    ) private view {
        _require(distributor.owner() == expectedOwner, "unexpected owner");
        _require(distributor.feeWallet() == expectedFeeWallet, "unexpected fee wallet");
        _require(distributor.feeBps() == 500, "unexpected fee rate");
        _require(distributor.usdt() == address(token), "unexpected USDT");

        (bool success, bytes memory data) =
            address(distributor).staticcall(abi.encodeWithSelector(USDTRewardDistributor.getRecipients.selector));
        _require(success, "recipient read failed");

        USDTRewardDistributor.Recipient[] memory actualRecipients =
            abi.decode(data, (USDTRewardDistributor.Recipient[]));
        _require(actualRecipients.length == expectedRecipients.length, "unexpected recipient count");
        for (uint256 i; i < actualRecipients.length; ++i) {
            _require(actualRecipients[i].account == expectedRecipients[i].account, "unexpected recipient account");
            _require(actualRecipients[i].shareBps == expectedRecipients[i].shareBps, "unexpected recipient share");
        }
    }

    function _initialRecipients() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](2);
        recipients[0] = USDTRewardDistributor.Recipient({account: address(3), shareBps: 5000});
        recipients[1] = USDTRewardDistributor.Recipient({account: address(4), shareBps: 5000});
    }

    function _updatedRecipients() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](3);
        recipients[0] = USDTRewardDistributor.Recipient({account: address(0x11), shareBps: 1000});
        recipients[1] = USDTRewardDistributor.Recipient({account: address(0x22), shareBps: 2000});
        recipients[2] = USDTRewardDistributor.Recipient({account: address(0x33), shareBps: 7000});
    }

    function _require(bool condition, string memory message) private pure {
        require(condition, message);
    }
}
