// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

interface Vm {
    function createSelectFork(string calldata rpcUrl, uint256 blockNumber) external returns (uint256 forkId);

    function envString(string calldata name) external returns (string memory value);

    function prank(address msgSender) external;
}

interface IERC20Metadata {
    function balanceOf(address account) external view returns (uint256);

    function decimals() external view returns (uint8);
}

/// @notice Pinned compatibility checks for the two production tokens in the OpenSpec change.
/// @dev Run only against read-only forks. Transfer probes use each token's on-chain owner and zero amount.
contract USDTDeploymentTest {
    address private constant ETHEREUM_USDT = 0xdAC17F958D2ee523a2206206994597C13D831ec7;
    address private constant ARBITRUM_USDT0 = 0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9;

    uint256 private constant ETHEREUM_CHAIN_ID = 1;
    uint256 private constant ARBITRUM_CHAIN_ID = 42161;
    uint256 private constant ETHEREUM_FORK_BLOCK = 26_469_961;
    uint256 private constant ARBITRUM_FORK_BLOCK = 526_486_711;
    address private constant ETHEREUM_TRANSFER_CALLER = 0xC6CDE7C39eB2f0F0095F41570af89eFC2C1Ea828;
    address private constant ARBITRUM_TRANSFER_CALLER = 0x4DFF9b5b0143E642a3F63a5bcf2d1C328e600bf8;
    bytes4 private constant TRANSFER_SELECTOR = bytes4(keccak256("transfer(address,uint256)"));

    Vm private constant vm = Vm(address(uint160(uint256(keccak256("hevm cheat code")))));

    function testEthereumUSDTDeploymentCodeAndDecimals() external {
        _selectEthereumFork();

        _assertEq(block.chainid, ETHEREUM_CHAIN_ID, "wrong Ethereum chain");
        _assertTrue(ETHEREUM_USDT.code.length > 0, "Ethereum USDT has no code");
        _assertEq(IERC20Metadata(ETHEREUM_USDT).decimals(), 6, "wrong Ethereum USDT decimals");
    }

    function testArbitrumUSDT0DeploymentCodeAndDecimals() external {
        _selectArbitrumFork();

        _assertEq(block.chainid, ARBITRUM_CHAIN_ID, "wrong Arbitrum chain");
        _assertTrue(ARBITRUM_USDT0.code.length > 0, "Arbitrum USDT0 has no code");
        _assertEq(IERC20Metadata(ARBITRUM_USDT0).decimals(), 6, "wrong Arbitrum USDT0 decimals");
    }

    function testEthereumUSDTTransferSucceedsWithoutBooleanReturnValue() external {
        _selectEthereumFork();
        uint256 balanceBefore = IERC20Metadata(ETHEREUM_USDT).balanceOf(ETHEREUM_TRANSFER_CALLER);
        vm.prank(ETHEREUM_TRANSFER_CALLER);

        (bool success, bytes memory returnData) =
            ETHEREUM_USDT.call(abi.encodeWithSelector(TRANSFER_SELECTOR, ETHEREUM_TRANSFER_CALLER, 0));

        _assertTrue(success, "Ethereum USDT transfer failed");
        _assertEq(returnData.length, 0, "Ethereum USDT returned data");
        _assertEq(
            IERC20Metadata(ETHEREUM_USDT).balanceOf(ETHEREUM_TRANSFER_CALLER),
            balanceBefore,
            "zero transfer changed balance"
        );
    }

    function testArbitrumUSDT0TransferSucceeds() external {
        _selectArbitrumFork();
        uint256 balanceBefore = IERC20Metadata(ARBITRUM_USDT0).balanceOf(ARBITRUM_TRANSFER_CALLER);
        vm.prank(ARBITRUM_TRANSFER_CALLER);

        (bool success, bytes memory returnData) =
            ARBITRUM_USDT0.call(abi.encodeWithSelector(TRANSFER_SELECTOR, ARBITRUM_TRANSFER_CALLER, 0));

        _assertTrue(success, "Arbitrum USDT0 transfer failed");
        _assertEq(returnData.length, 32, "Arbitrum USDT0 returned unexpected data");
        _assertTrue(abi.decode(returnData, (bool)), "Arbitrum USDT0 reported transfer failure");
        _assertEq(
            IERC20Metadata(ARBITRUM_USDT0).balanceOf(ARBITRUM_TRANSFER_CALLER),
            balanceBefore,
            "zero transfer changed balance"
        );
    }

    function _selectEthereumFork() private {
        vm.createSelectFork(vm.envString("ETHEREUM_RPC_URL"), ETHEREUM_FORK_BLOCK);
    }

    function _selectArbitrumFork() private {
        vm.createSelectFork(vm.envString("ARBITRUM_RPC_URL"), ARBITRUM_FORK_BLOCK);
    }

    function _assertTrue(bool condition, string memory message) private pure {
        require(condition, message);
    }

    function _assertEq(uint256 actual, uint256 expected, string memory message) private pure {
        require(actual == expected, message);
    }
}
