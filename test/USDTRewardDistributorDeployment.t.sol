// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

import {USDTRewardDistributor} from "../contracts/USDTRewardDistributor.sol";

interface Vm {
    function computeCreateAddress(address deployer, uint64 nonce) external returns (address);
    function getNonce(address account) external returns (uint64);
}

contract MockUSDT {}

contract InvalidSelfAddressFactory {
    Vm private constant VM = Vm(address(uint160(uint256(keccak256("hevm cheat code")))));

    function deployWithSelfAddress(address token, uint8 position) external returns (USDTRewardDistributor) {
        address distributor = VM.computeCreateAddress(address(this), VM.getNonce(address(this)));
        USDTRewardDistributor.Recipient[] memory initialRecipients = new USDTRewardDistributor.Recipient[](2);
        initialRecipients[0] = USDTRewardDistributor.Recipient({account: address(3), shareBps: 5000});
        initialRecipients[1] = USDTRewardDistributor.Recipient({account: address(4), shareBps: 5000});

        address initialOwner = position == 0 ? distributor : address(1);
        address initialFeeWallet = position == 1 ? distributor : address(2);
        if (position == 2) {
            initialRecipients[0].account = distributor;
        }

        return new USDTRewardDistributor(token, 0, initialOwner, initialFeeWallet, initialRecipients);
    }
}

contract USDTRewardDistributorDeploymentTest {
    MockUSDT private token = new MockUSDT();

    function testDeploysWithCompleteInitialConfiguration() external {
        USDTRewardDistributor.Recipient[] memory initialRecipients = _twoRecipients();
        address initialOwner = address(0xA11CE);
        address initialFeeWallet = address(0xFEE);

        USDTRewardDistributor distributor =
            new USDTRewardDistributor(address(token), 500, initialOwner, initialFeeWallet, initialRecipients);

        _require(distributor.usdt() == address(token), "unexpected USDT address");
        _require(distributor.feeBps() == 500, "unexpected fee rate");
        _require(distributor.owner() == initialOwner, "unexpected owner");
        _require(distributor.feeWallet() == initialFeeWallet, "unexpected fee wallet");
    }

    function testDeploymentStateIsIndependent() external {
        USDTRewardDistributor.Recipient[] memory firstRecipients = _twoRecipients();
        USDTRewardDistributor.Recipient[] memory secondRecipients = _twoRecipients();

        USDTRewardDistributor first =
            new USDTRewardDistributor(address(token), 0, address(0xA11CE), address(0xFEE), firstRecipients);
        USDTRewardDistributor second =
            new USDTRewardDistributor(address(token), 9999, address(0xB0B), address(0xCAFE), secondRecipients);

        _require(address(first) != address(second), "deployments must differ");
        _require(first.owner() == address(0xA11CE), "first owner changed");
        _require(first.feeBps() == 0, "first fee changed");
        _require(second.owner() == address(0xB0B), "second owner changed");
        _require(second.feeBps() == 9999, "second fee changed");
    }

    function testAcceptsFeeBounds() external {
        USDTRewardDistributor.Recipient[] memory initialRecipients = _twoRecipients();
        new USDTRewardDistributor(address(token), 0, address(1), address(2), initialRecipients);
        new USDTRewardDistributor(address(token), 9999, address(1), address(2), initialRecipients);
    }

    function testAllowsOwnerAndFeeWalletAsRecipients() external {
        USDTRewardDistributor.Recipient[] memory initialRecipients = _twoRecipients();
        new USDTRewardDistributor(address(token), 0, address(3), address(4), initialRecipients);
    }

    function testRejectsInvalidTokenAddresses() external {
        _require(_deploymentReverts(address(0), 0, address(1), address(2), _twoRecipients()), "zero token accepted");
        _require(_deploymentReverts(address(0x1234), 0, address(1), address(2), _twoRecipients()), "EOA token accepted");
    }

    function testRejectsInvalidOwnerAndFeeWalletAddresses() external {
        _require(_deploymentReverts(address(token), 0, address(0), address(2), _twoRecipients()), "zero owner accepted");
        _require(
            _deploymentReverts(address(token), 0, address(1), address(0), _twoRecipients()), "zero fee wallet accepted"
        );
    }

    function testRejectsDistributorAsOwnerFeeWalletOrRecipient() external {
        InvalidSelfAddressFactory factory = new InvalidSelfAddressFactory();

        _require(_selfAddressDeploymentReverts(factory, 0), "distributor owner accepted");
        _require(_selfAddressDeploymentReverts(factory, 1), "distributor fee wallet accepted");
        _require(_selfAddressDeploymentReverts(factory, 2), "distributor recipient accepted");
    }

    function testRejectsOutOfRangeFee() external {
        _require(
            _deploymentReverts(address(token), 10_000, address(1), address(2), _twoRecipients()), "fee 10000 accepted"
        );
        _require(
            _deploymentReverts(address(token), 10_001, address(1), address(2), _twoRecipients()), "fee 10001 accepted"
        );
    }

    function testRejectsInvalidRecipientCounts() external {
        USDTRewardDistributor.Recipient[] memory empty = new USDTRewardDistributor.Recipient[](0);
        _require(_deploymentReverts(address(token), 0, address(1), address(2), empty), "empty list accepted");

        USDTRewardDistributor.Recipient[] memory tooMany = new USDTRewardDistributor.Recipient[](101);
        for (uint256 i; i < 101; ++i) {
            tooMany[i] = USDTRewardDistributor.Recipient({account: address(uint160(i + 1)), shareBps: 100});
        }
        _require(
            _deploymentReverts(address(token), 0, address(0xA11CE), address(0xFEE), tooMany), "101 recipients accepted"
        );
    }

    function testAcceptsMaximumRecipientCount() external {
        USDTRewardDistributor.Recipient[] memory maximum = new USDTRewardDistributor.Recipient[](100);
        for (uint256 i; i < 100; ++i) {
            maximum[i] = USDTRewardDistributor.Recipient({account: address(uint160(i + 1)), shareBps: 100});
        }
        new USDTRewardDistributor(address(token), 0, address(0xA11CE), address(0xFEE), maximum);
    }

    function testRejectsInvalidRecipientEntriesAndTotals() external {
        USDTRewardDistributor.Recipient[] memory duplicate = new USDTRewardDistributor.Recipient[](2);
        duplicate[0] = USDTRewardDistributor.Recipient({account: address(3), shareBps: 5000});
        duplicate[1] = USDTRewardDistributor.Recipient({account: address(3), shareBps: 5000});
        _require(
            _deploymentReverts(address(token), 0, address(1), address(2), duplicate), "duplicate recipient accepted"
        );

        USDTRewardDistributor.Recipient[] memory zeroRecipient = _twoRecipients();
        zeroRecipient[0].account = address(0);
        _require(
            _deploymentReverts(address(token), 0, address(1), address(2), zeroRecipient), "zero recipient accepted"
        );

        USDTRewardDistributor.Recipient[] memory zeroShare = _twoRecipients();
        zeroShare[0].shareBps = 0;
        _require(_deploymentReverts(address(token), 0, address(1), address(2), zeroShare), "zero share accepted");

        USDTRewardDistributor.Recipient[] memory lowTotal = _twoRecipients();
        lowTotal[1].shareBps = 4999;
        _require(_deploymentReverts(address(token), 0, address(1), address(2), lowTotal), "low total accepted");

        USDTRewardDistributor.Recipient[] memory highTotal = _twoRecipients();
        highTotal[1].shareBps = 5001;
        _require(_deploymentReverts(address(token), 0, address(1), address(2), highTotal), "high total accepted");
    }

    function _deploymentReverts(
        address initialUSDT,
        uint256 initialFeeBps,
        address initialOwner,
        address initialFeeWallet,
        USDTRewardDistributor.Recipient[] memory initialRecipients
    ) private returns (bool) {
        try new USDTRewardDistributor(initialUSDT, initialFeeBps, initialOwner, initialFeeWallet, initialRecipients)
        returns (USDTRewardDistributor) {
            return false;
        } catch {
            return true;
        }
    }

    function _selfAddressDeploymentReverts(InvalidSelfAddressFactory factory, uint8 position) private returns (bool) {
        try factory.deployWithSelfAddress(address(token), position) returns (USDTRewardDistributor) {
            return false;
        } catch {
            return true;
        }
    }

    function _twoRecipients() private pure returns (USDTRewardDistributor.Recipient[] memory recipients) {
        recipients = new USDTRewardDistributor.Recipient[](2);
        recipients[0] = USDTRewardDistributor.Recipient({account: address(3), shareBps: 5000});
        recipients[1] = USDTRewardDistributor.Recipient({account: address(4), shareBps: 5000});
    }

    function _require(bool condition, string memory message) private pure {
        require(condition, message);
    }
}
