// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

/// @notice Holds the validated, mutable configuration for a USDT reward distributor.
/// @dev Payout operations are added in later tasks.
contract USDTRewardDistributor {
    // Represents 100% in basis points (1 BPS = 0.01%) for validating fees and recipient shares.
    uint256 internal constant BPS_DENOMINATOR = 10_000;
    uint256 internal constant MAX_RECIPIENTS = 100;

    struct Recipient {
        address account;
        uint16 shareBps;
    }

    error InvalidToken(address token);
    error InvalidOwner(address owner);
    error InvalidFeeWallet(address feeWallet);
    error InvalidFeeBps(uint256 feeBps);
    error InvalidRecipientCount(uint256 count);
    error InvalidRecipient(address recipient);
    error DuplicateRecipient(address recipient);
    error InvalidRecipientShare(uint256 shareBps);
    error InvalidShareTotal(uint256 totalBps);

    address public immutable usdt;
    uint16 public immutable feeBps;
    address public owner;
    address public feeWallet;

    Recipient[] internal recipients;

    modifier onlyOwner() {
        require(msg.sender == owner, "only owner");
        _;
    }

    constructor(
        address initialUSDT,
        uint256 initialFeeBps,
        address initialOwner,
        address initialFeeWallet,
        Recipient[] memory initialRecipients
    ) {
        if (initialUSDT == address(0) || initialUSDT.code.length == 0) {
            revert InvalidToken(initialUSDT);
        }
        if (initialOwner == address(0) || initialOwner == address(this)) {
            revert InvalidOwner(initialOwner);
        }
        if (initialFeeWallet == address(0) || initialFeeWallet == address(this)) {
            revert InvalidFeeWallet(initialFeeWallet);
        }
        if (initialFeeBps >= BPS_DENOMINATOR) {
            revert InvalidFeeBps(initialFeeBps);
        }

        _validateAndStoreRecipients(initialRecipients);

        usdt = initialUSDT;
        feeBps = uint16(initialFeeBps);
        owner = initialOwner;
        feeWallet = initialFeeWallet;
    }

    function _validateAndStoreRecipients(Recipient[] memory initialRecipients) private {
        uint256 length = initialRecipients.length;
        if (length == 0 || length > MAX_RECIPIENTS) {
            revert InvalidRecipientCount(length);
        }

        uint256 totalBps;
        for (uint256 i; i < length; ++i) {
            Recipient memory recipient = initialRecipients[i];
            if (recipient.account == address(0) || recipient.account == address(this)) {
                revert InvalidRecipient(recipient.account);
            }
            if (recipient.shareBps == 0) {
                revert InvalidRecipientShare(recipient.shareBps);
            }

            for (uint256 j; j < i; ++j) {
                if (initialRecipients[j].account == recipient.account) {
                    revert DuplicateRecipient(recipient.account);
                }
            }

            totalBps += recipient.shareBps;
        }

        if (totalBps != BPS_DENOMINATOR) {
            revert InvalidShareTotal(totalBps);
        }

        delete recipients;
        for (uint256 i; i < length; ++i) {
            recipients.push(initialRecipients[i]);
        }
    }

    /// @notice Allows the owner to fully replace all recipients and their shares.
    /// @dev Validates 1..MAX_RECIPIENTS unique addresses, excluding zero and this contract,
    /// with positive shares totaling 10,000 BPS (100%), before overwriting the existing list.
    function replaceRecipients(Recipient[] memory newRecipients) external onlyOwner {
        _validateAndStoreRecipients(newRecipients);
    }

    function getRecipients() external view returns (Recipient[] memory) {
        return recipients;
    }

    function transferOwnership(address newOwner) external onlyOwner {
        if (newOwner == address(0) || newOwner == address(this)) {
            revert InvalidOwner(newOwner);
        }

        owner = newOwner;
    }

    function setFeeWallet(address newFeeWallet) external onlyOwner {
        if (newFeeWallet == address(0) || newFeeWallet == address(this)) {
            revert InvalidFeeWallet(newFeeWallet);
        }

        feeWallet = newFeeWallet;
    }
}
