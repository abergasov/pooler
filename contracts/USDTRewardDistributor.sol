// SPDX-License-Identifier: MIT
pragma solidity 0.8.28;

interface IERC20Balance {
    function balanceOf(address account) external view returns (uint256);
}

/// @notice Holds the validated, mutable configuration for a USDT reward distributor.
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
    error InvalidWithdrawalDestination(address destination);
    error EmptyBalance();
    error TokenTransferFailed(address recipient, uint256 amount);
    error RemainingBalance(uint256 balance);
    error ReentrantCall();

    address public immutable usdt;
    uint16 public immutable feeBps;
    address public owner;
    address public feeWallet;
    bool private locked;

    Recipient[] internal recipients;

    modifier onlyOwner() {
        require(msg.sender == owner, "only owner");
        _;
    }

    modifier nonReentrant() {
        if (locked) {
            revert ReentrantCall();
        }

        locked = true;
        _;
        locked = false;
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
    function replaceRecipients(Recipient[] memory newRecipients) external onlyOwner nonReentrant {
        _validateAndStoreRecipients(newRecipients);
    }

    function getRecipients() external view returns (Recipient[] memory) {
        return recipients;
    }

    function distribute() external onlyOwner nonReentrant {
        uint256 balance = IERC20Balance(usdt).balanceOf(address(this));
        if (balance == 0) {
            revert EmptyBalance();
        }

        uint256 fee = (balance * feeBps) / BPS_DENOMINATOR;
        uint256 participantPool = balance - fee;
        uint256 paidToRecipients;
        uint256 length = recipients.length;

        for (uint256 i; i < length; ++i) {
            Recipient memory recipient = recipients[i];
            uint256 payout = (participantPool * recipient.shareBps) / BPS_DENOMINATOR;
            paidToRecipients += payout;
            if (payout != 0) {
                _transferUSDT(recipient.account, payout);
            }
        }

        uint256 feePayout = balance - paidToRecipients;
        if (feePayout != 0) {
            _transferUSDT(feeWallet, feePayout);
        }

        uint256 remainingBalance = IERC20Balance(usdt).balanceOf(address(this));
        if (remainingBalance != 0) {
            revert RemainingBalance(remainingBalance);
        }
    }

    function withdrawUSDT(address destination) external onlyOwner nonReentrant {
        uint256 balance = IERC20Balance(usdt).balanceOf(address(this));
        if (balance == 0) {
            revert EmptyBalance();
        }
        if (destination == address(0) || destination == address(this)) {
            revert InvalidWithdrawalDestination(destination);
        }

        _transferUSDT(destination, balance);

        uint256 remainingBalance = IERC20Balance(usdt).balanceOf(address(this));
        if (remainingBalance != 0) {
            revert RemainingBalance(remainingBalance);
        }
    }

    function transferOwnership(address newOwner) external onlyOwner nonReentrant {
        if (newOwner == address(0) || newOwner == address(this)) {
            revert InvalidOwner(newOwner);
        }

        owner = newOwner;
    }

    function setFeeWallet(address newFeeWallet) external onlyOwner nonReentrant {
        if (newFeeWallet == address(0) || newFeeWallet == address(this)) {
            revert InvalidFeeWallet(newFeeWallet);
        }

        feeWallet = newFeeWallet;
    }

    function _transferUSDT(address recipient, uint256 amount) private {
        (bool success, bytes memory returnData) =
            usdt.call(abi.encodeWithSelector(bytes4(keccak256("transfer(address,uint256)")), recipient, amount));
        if (!success || (returnData.length != 0 && (returnData.length != 32 || !abi.decode(returnData, (bool))))) {
            revert TokenTransferFailed(recipient, amount);
        }
    }
}
