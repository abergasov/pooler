// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"context"
	"errors"
	"math/big"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
	_ = time.Tick
	_ = context.Background
)

// IERC20BalanceMetaData contains all meta data concerning the IERC20Balance contract.
var IERC20BalanceMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"balanceOf\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"}]",
}

// IERC20BalanceABI is the input ABI used to generate the binding from.
// Deprecated: Use IERC20BalanceMetaData.ABI instead.
var IERC20BalanceABI = IERC20BalanceMetaData.ABI

// IERC20Balance is an auto generated Go binding around an Ethereum contract.
type IERC20Balance struct {
	IERC20BalanceCaller     // Read-only binding to the contract
	IERC20BalanceTransactor // Write-only binding to the contract
	IERC20BalanceFilterer   // Log filterer for contract events
}

// IERC20BalanceCaller is an auto generated read-only Go binding around an Ethereum contract.
type IERC20BalanceCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20BalanceTransactor is an auto generated write-only Go binding around an Ethereum contract.
type IERC20BalanceTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20BalanceFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type IERC20BalanceFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// IERC20BalanceSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type IERC20BalanceSession struct {
	Contract     *IERC20Balance    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// IERC20BalanceCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type IERC20BalanceCallerSession struct {
	Contract *IERC20BalanceCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// IERC20BalanceTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type IERC20BalanceTransactorSession struct {
	Contract     *IERC20BalanceTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// IERC20BalanceRaw is an auto generated low-level Go binding around an Ethereum contract.
type IERC20BalanceRaw struct {
	Contract *IERC20Balance // Generic contract binding to access the raw methods on
}

// IERC20BalanceCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type IERC20BalanceCallerRaw struct {
	Contract *IERC20BalanceCaller // Generic read-only contract binding to access the raw methods on
}

// IERC20BalanceTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type IERC20BalanceTransactorRaw struct {
	Contract *IERC20BalanceTransactor // Generic write-only contract binding to access the raw methods on
}

// NewIERC20Balance creates a new instance of IERC20Balance, bound to a specific deployed contract.
func NewIERC20Balance(address common.Address, backend bind.ContractBackend) (*IERC20Balance, error) {
	contract, err := bindIERC20Balance(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &IERC20Balance{IERC20BalanceCaller: IERC20BalanceCaller{contract: contract}, IERC20BalanceTransactor: IERC20BalanceTransactor{contract: contract}, IERC20BalanceFilterer: IERC20BalanceFilterer{contract: contract}}, nil
}

// NewIERC20BalanceCaller creates a new read-only instance of IERC20Balance, bound to a specific deployed contract.
func NewIERC20BalanceCaller(address common.Address, caller bind.ContractCaller) (*IERC20BalanceCaller, error) {
	contract, err := bindIERC20Balance(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &IERC20BalanceCaller{contract: contract}, nil
}

// NewIERC20BalanceTransactor creates a new write-only instance of IERC20Balance, bound to a specific deployed contract.
func NewIERC20BalanceTransactor(address common.Address, transactor bind.ContractTransactor) (*IERC20BalanceTransactor, error) {
	contract, err := bindIERC20Balance(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &IERC20BalanceTransactor{contract: contract}, nil
}

// NewIERC20BalanceFilterer creates a new log filterer instance of IERC20Balance, bound to a specific deployed contract.
func NewIERC20BalanceFilterer(address common.Address, filterer bind.ContractFilterer) (*IERC20BalanceFilterer, error) {
	contract, err := bindIERC20Balance(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &IERC20BalanceFilterer{contract: contract}, nil
}

// bindIERC20Balance binds a generic wrapper to an already deployed contract.
func bindIERC20Balance(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := IERC20BalanceMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IERC20Balance *IERC20BalanceRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IERC20Balance.Contract.IERC20BalanceCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IERC20Balance *IERC20BalanceRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IERC20Balance.Contract.IERC20BalanceTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IERC20Balance *IERC20BalanceRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IERC20Balance.Contract.IERC20BalanceTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_IERC20Balance *IERC20BalanceCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _IERC20Balance.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_IERC20Balance *IERC20BalanceTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _IERC20Balance.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_IERC20Balance *IERC20BalanceTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _IERC20Balance.Contract.contract.Transact(opts, method, params...)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_IERC20Balance *IERC20BalanceCaller) BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error) {
	var out []interface{}
	err := _IERC20Balance.contract.Call(opts, &out, "balanceOf", account)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_IERC20Balance *IERC20BalanceSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _IERC20Balance.Contract.BalanceOf(&_IERC20Balance.CallOpts, account)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address account) view returns(uint256)
func (_IERC20Balance *IERC20BalanceCallerSession) BalanceOf(account common.Address) (*big.Int, error) {
	return _IERC20Balance.Contract.BalanceOf(&_IERC20Balance.CallOpts, account)
}
