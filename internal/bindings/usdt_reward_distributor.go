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

// USDTRewardDistributorRecipient is an auto generated low-level Go binding around an user-defined struct.
type USDTRewardDistributorRecipient struct {
	Account  common.Address
	ShareBps uint16
}

// USDTRewardDistributorMetaData contains all meta data concerning the USDTRewardDistributor contract.
var USDTRewardDistributorMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"initialUSDT\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"initialFeeBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initialOwner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"initialFeeWallet\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"initialRecipients\",\"type\":\"tuple[]\",\"internalType\":\"structUSDTRewardDistributor.Recipient[]\",\"components\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"shareBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"distribute\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"feeBps\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"feeWallet\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getRecipients\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple[]\",\"internalType\":\"structUSDTRewardDistributor.Recipient[]\",\"components\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"shareBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"replaceRecipients\",\"inputs\":[{\"name\":\"newRecipients\",\"type\":\"tuple[]\",\"internalType\":\"structUSDTRewardDistributor.Recipient[]\",\"components\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"shareBps\",\"type\":\"uint16\",\"internalType\":\"uint16\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setFeeWallet\",\"inputs\":[{\"name\":\"newFeeWallet\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"usdt\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"withdrawUSDT\",\"inputs\":[{\"name\":\"destination\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"error\",\"name\":\"DuplicateRecipient\",\"inputs\":[{\"name\":\"recipient\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"EmptyBalance\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidFeeBps\",\"inputs\":[{\"name\":\"feeBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidFeeWallet\",\"inputs\":[{\"name\":\"feeWallet\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"InvalidOwner\",\"inputs\":[{\"name\":\"owner\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"InvalidRecipient\",\"inputs\":[{\"name\":\"recipient\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"InvalidRecipientCount\",\"inputs\":[{\"name\":\"count\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidRecipientShare\",\"inputs\":[{\"name\":\"shareBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidShareTotal\",\"inputs\":[{\"name\":\"totalBps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"InvalidToken\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"InvalidWithdrawalDestination\",\"inputs\":[{\"name\":\"destination\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"RemainingBalance\",\"inputs\":[{\"name\":\"balance\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"TokenTransferFailed\",\"inputs\":[{\"name\":\"recipient\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}]",
	Bin: "0x60c060405234801561000f575f5ffd5b506040516128ab3803806128ab83398181016040528101906100319190610976565b5f73ffffffffffffffffffffffffffffffffffffffff168573ffffffffffffffffffffffffffffffffffffffff16148061008157505f8573ffffffffffffffffffffffffffffffffffffffff163b145b156100c357846040517f961c9a4f0000000000000000000000000000000000000000000000000000000081526004016100ba9190610a18565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16148061012857503073ffffffffffffffffffffffffffffffffffffffff168373ffffffffffffffffffffffffffffffffffffffff16145b1561016a57826040517fb20f76e30000000000000000000000000000000000000000000000000000000081526004016101619190610a18565b60405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff1614806101cf57503073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16145b1561021157816040517fb63bc9cc0000000000000000000000000000000000000000000000000000000081526004016102089190610a18565b60405180910390fd5b612710841061025757836040517f3329f93200000000000000000000000000000000000000000000000000000000815260040161024e9190610a40565b60405180910390fd5b6102668161033360201b60201c565b8473ffffffffffffffffffffffffffffffffffffffff1660808173ffffffffffffffffffffffffffffffffffffffff16815250508361ffff1660a08161ffff1681525050825f5f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055508160015f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505050505050610b38565b5f815190505f8114806103465750606481115b1561038857806040517f37d33faf00000000000000000000000000000000000000000000000000000000815260040161037f9190610a40565b60405180910390fd5b5f5f5b8281101561057c575f8482815181106103a7576103a6610a59565b5b602002602001015190505f73ffffffffffffffffffffffffffffffffffffffff16815f015173ffffffffffffffffffffffffffffffffffffffff16148061041c57503073ffffffffffffffffffffffffffffffffffffffff16815f015173ffffffffffffffffffffffffffffffffffffffff16145b1561046157805f01516040517f17858bbe0000000000000000000000000000000000000000000000000000000081526004016104589190610a18565b60405180910390fd5b5f816020015161ffff16036104b15780602001516040517f373b8c910000000000000000000000000000000000000000000000000000000081526004016104a89190610abf565b60405180910390fd5b5f5b8281101561055957815f015173ffffffffffffffffffffffffffffffffffffffff168682815181106104e8576104e7610a59565b5b60200260200101515f015173ffffffffffffffffffffffffffffffffffffffff160361054e57815f01516040517f3e9185790000000000000000000000000000000000000000000000000000000081526004016105459190610a18565b60405180910390fd5b8060010190506104b3565b50806020015161ffff168361056e9190610b05565b92505080600101905061038b565b5061271081146105c357806040517fad533c2b0000000000000000000000000000000000000000000000000000000081526004016105ba9190610a40565b60405180910390fd5b60025f6105d09190610692565b5f5b8281101561068c5760028482815181106105ef576105ee610a59565b5b6020026020010151908060018154018082558091505060019003905f5260205f20015f909190919091505f820151815f015f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506020820151815f0160146101000a81548161ffff021916908361ffff16021790555050508060010190506105d2565b50505050565b5080545f8255905f5260205f20908101906106ad91906106b0565b50565b5b808211156106fc575f5f82015f6101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690555f820160146101000a81549061ffff0219169055506001016106b1565b5090565b5f604051905090565b5f5ffd5b5f5ffd5b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61073a82610711565b9050919050565b61074a81610730565b8114610754575f5ffd5b50565b5f8151905061076581610741565b92915050565b5f819050919050565b61077d8161076b565b8114610787575f5ffd5b50565b5f8151905061079881610774565b92915050565b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b6107e8826107a2565b810181811067ffffffffffffffff82111715610807576108066107b2565b5b80604052505050565b5f610819610700565b905061082582826107df565b919050565b5f67ffffffffffffffff821115610844576108436107b2565b5b602082029050602081019050919050565b5f5ffd5b5f5ffd5b5f61ffff82169050919050565b6108738161085d565b811461087d575f5ffd5b50565b5f8151905061088e8161086a565b92915050565b5f604082840312156108a9576108a8610859565b5b6108b36040610810565b90505f6108c284828501610757565b5f8301525060206108d584828501610880565b60208301525092915050565b5f6108f36108ee8461082a565b610810565b9050808382526020820190506040840283018581111561091657610915610855565b5b835b8181101561093f578061092b8882610894565b845260208401935050604081019050610918565b5050509392505050565b5f82601f83011261095d5761095c61079e565b5b815161096d8482602086016108e1565b91505092915050565b5f5f5f5f5f60a0868803121561098f5761098e610709565b5b5f61099c88828901610757565b95505060206109ad8882890161078a565b94505060406109be88828901610757565b93505060606109cf88828901610757565b925050608086015167ffffffffffffffff8111156109f0576109ef61070d565b5b6109fc88828901610949565b9150509295509295909350565b610a1281610730565b82525050565b5f602082019050610a2b5f830184610a09565b92915050565b610a3a8161076b565b82525050565b5f602082019050610a535f830184610a31565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b5f819050919050565b5f610aa9610aa4610a9f8461085d565b610a86565b61076b565b9050919050565b610ab981610a8f565b82525050565b5f602082019050610ad25f830184610ab0565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f610b0f8261076b565b9150610b1a8361076b565b9250828201905080821115610b3257610b31610ad8565b5b92915050565b60805160a051611d28610b835f395f81816101b20152610b2a01525f81816101d6015281816102ea0152818161046f01528181610a5301528181610cbe0152610fd80152611d285ff3fe608060405234801561000f575f5ffd5b506004361061009c575f3560e01c8063a6e845d111610064578063a6e845d114610132578063d78d610b1461014e578063e4fc6b6d1461016c578063f25f4b5614610176578063f2fde38b146101945761009c565b806324a9d853146100a05780632f48ab7d146100be57806359db80f8146100dc5780638da5cb5b146100f857806390d49b9d14610116575b5f5ffd5b6100a86101b0565b6040516100b59190611559565b60405180910390f35b6100c66101d4565b6040516100d391906115b1565b60405180910390f35b6100f660048036038101906100f19190611605565b6101f8565b005b61010061056a565b60405161010d91906115b1565b60405180910390f35b610130600480360381019061012b9190611605565b61058e565b005b61014c600480360381019061014791906117fb565b610781565b005b610156610896565b6040516101639190611935565b60405180910390f35b610174610961565b005b61017e610dbd565b60405161018b91906115b1565b60405180910390f35b6101ae60048036038101906101a99190611605565b610de2565b005b7f000000000000000000000000000000000000000000000000000000000000000081565b7f000000000000000000000000000000000000000000000000000000000000000081565b5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610286576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161027d906119af565b60405180910390fd5b600160149054906101000a900460ff16156102cd576040517f37ed32e800000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60018060146101000a81548160ff0219169083151502179055505f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b815260040161034191906115b1565b602060405180830381865afa15801561035c573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906103809190611a00565b90505f81036103bb576040517f2313b4b600000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f73ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16148061042057503073ffffffffffffffffffffffffffffffffffffffff168273ffffffffffffffffffffffffffffffffffffffff16145b1561046257816040517ffd96582600000000000000000000000000000000000000000000000000000000815260040161045991906115b1565b60405180910390fd5b61046c8282610fd4565b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b81526004016104c691906115b1565b602060405180830381865afa1580156104e1573d5f5f3e3d5ffd5b505050506040513d601f19601f820116820180604052508101906105059190611a00565b90505f811461054b57806040517f3939d0b60000000000000000000000000000000000000000000000000000000081526004016105429190611a3a565b60405180910390fd5b50505f600160146101000a81548160ff02191690831515021790555050565b5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461061c576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610613906119af565b60405180910390fd5b600160149054906101000a900460ff1615610663576040517f37ed32e800000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60018060146101000a81548160ff0219169083151502179055505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff1614806106e257503073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16145b1561072457806040517fb63bc9cc00000000000000000000000000000000000000000000000000000000815260040161071b91906115b1565b60405180910390fd5b8060015f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505f600160146101000a81548160ff02191690831515021790555050565b5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff161461080f576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610806906119af565b60405180910390fd5b600160149054906101000a900460ff1615610856576040517f37ed32e800000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60018060146101000a81548160ff02191690831515021790555061087981611170565b5f600160146101000a81548160ff02191690831515021790555050565b60606002805480602002602001604051908101604052809291908181526020015f905b82821015610958578382905f5260205f20016040518060400160405290815f82015f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020015f820160149054906101000a900461ffff1661ffff1661ffff1681525050815260200190600101906108b9565b50505050905090565b5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff16146109ef576040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016109e6906119af565b60405180910390fd5b600160149054906101000a900460ff1615610a36576040517f37ed32e800000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60018060146101000a81548160ff0219169083151502179055505f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b8152600401610aaa91906115b1565b602060405180830381865afa158015610ac5573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610ae99190611a00565b90505f8103610b24576040517f2313b4b600000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b5f6127107f000000000000000000000000000000000000000000000000000000000000000061ffff1683610b589190611a80565b610b629190611aee565b90505f8183610b719190611b1e565b90505f5f60028054905090505f5b81811015610c78575f60028281548110610b9c57610b9b611b51565b5b905f5260205f20016040518060400160405290815f82015f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff1681526020015f820160149054906101000a900461ffff1661ffff1661ffff168152505090505f612710826020015161ffff1687610c3c9190611a80565b610c469190611aee565b90508085610c549190611b7e565b94505f8114610c6b57610c6a825f015182610fd4565b5b5050806001019050610b7f565b505f8286610c869190611b1e565b90505f8114610cbb57610cba60015f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1682610fd4565b5b5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166370a08231306040518263ffffffff1660e01b8152600401610d1591906115b1565b602060405180830381865afa158015610d30573d5f5f3e3d5ffd5b505050506040513d601f19601f82011682018060405250810190610d549190611a00565b90505f8114610d9a57806040517f3939d0b6000000000000000000000000000000000000000000000000000000008152600401610d919190611a3a565b60405180910390fd5b505050505050505f600160146101000a81548160ff021916908315150217905550565b60015f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1681565b5f5f9054906101000a900473ffffffffffffffffffffffffffffffffffffffff1673ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff1614610e70576040517f08c379a0000000000000000000000000000000000000000000000000000000008152600401610e67906119af565b60405180910390fd5b600160149054906101000a900460ff1615610eb7576040517f37ed32e800000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b60018060146101000a81548160ff0219169083151502179055505f73ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff161480610f3657503073ffffffffffffffffffffffffffffffffffffffff168173ffffffffffffffffffffffffffffffffffffffff16145b15610f7857806040517fb20f76e3000000000000000000000000000000000000000000000000000000008152600401610f6f91906115b1565b60405180910390fd5b805f5f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055505f600160146101000a81548160ff02191690831515021790555050565b5f5f7f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff167fa9059cbb2ab09eb219583f4a59a5d0623ade346d962bcd4e46b11da047c9049b8585604051602401611041929190611bb1565b604051602081830303815290604052907bffffffffffffffffffffffffffffffffffffffffffffffffffffffff19166020820180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff83818316178352505050506040516110ab9190611c2a565b5f604051808303815f865af19150503d805f81146110e4576040519150601f19603f3d011682016040523d82523d5f602084013e6110e9565b606091505b509150915081158061112657505f8151141580156111255750602081511415806111245750808060200190518101906111229190611c75565b155b5b5b1561116a5783836040517f127744c0000000000000000000000000000000000000000000000000000000008152600401611161929190611bb1565b60405180910390fd5b50505050565b5f815190505f8114806111835750606481115b156111c557806040517f37d33faf0000000000000000000000000000000000000000000000000000000081526004016111bc9190611a3a565b60405180910390fd5b5f5f5b828110156113b9575f8482815181106111e4576111e3611b51565b5b602002602001015190505f73ffffffffffffffffffffffffffffffffffffffff16815f015173ffffffffffffffffffffffffffffffffffffffff16148061125957503073ffffffffffffffffffffffffffffffffffffffff16815f015173ffffffffffffffffffffffffffffffffffffffff16145b1561129e57805f01516040517f17858bbe00000000000000000000000000000000000000000000000000000000815260040161129591906115b1565b60405180910390fd5b5f816020015161ffff16036112ee5780602001516040517f373b8c910000000000000000000000000000000000000000000000000000000081526004016112e59190611cd9565b60405180910390fd5b5f5b8281101561139657815f015173ffffffffffffffffffffffffffffffffffffffff1686828151811061132557611324611b51565b5b60200260200101515f015173ffffffffffffffffffffffffffffffffffffffff160361138b57815f01516040517f3e91857900000000000000000000000000000000000000000000000000000000815260040161138291906115b1565b60405180910390fd5b8060010190506112f0565b50806020015161ffff16836113ab9190611b7e565b9250508060010190506111c8565b50612710811461140057806040517fad533c2b0000000000000000000000000000000000000000000000000000000081526004016113f79190611a3a565b60405180910390fd5b60025f61140d91906114cf565b5f5b828110156114c957600284828151811061142c5761142b611b51565b5b6020026020010151908060018154018082558091505060019003905f5260205f20015f909190919091505f820151815f015f6101000a81548173ffffffffffffffffffffffffffffffffffffffff021916908373ffffffffffffffffffffffffffffffffffffffff1602179055506020820151815f0160146101000a81548161ffff021916908361ffff160217905550505080600101905061140f565b50505050565b5080545f8255905f5260205f20908101906114ea91906114ed565b50565b5b80821115611539575f5f82015f6101000a81549073ffffffffffffffffffffffffffffffffffffffff02191690555f820160146101000a81549061ffff0219169055506001016114ee565b5090565b5f61ffff82169050919050565b6115538161153d565b82525050565b5f60208201905061156c5f83018461154a565b92915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61159b82611572565b9050919050565b6115ab81611591565b82525050565b5f6020820190506115c45f8301846115a2565b92915050565b5f604051905090565b5f5ffd5b5f5ffd5b6115e481611591565b81146115ee575f5ffd5b50565b5f813590506115ff816115db565b92915050565b5f6020828403121561161a576116196115d3565b5b5f611627848285016115f1565b91505092915050565b5f5ffd5b5f601f19601f8301169050919050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b61167a82611634565b810181811067ffffffffffffffff8211171561169957611698611644565b5b80604052505050565b5f6116ab6115ca565b90506116b78282611671565b919050565b5f67ffffffffffffffff8211156116d6576116d5611644565b5b602082029050602081019050919050565b5f5ffd5b5f5ffd5b6116f88161153d565b8114611702575f5ffd5b50565b5f81359050611713816116ef565b92915050565b5f6040828403121561172e5761172d6116eb565b5b61173860406116a2565b90505f611747848285016115f1565b5f83015250602061175a84828501611705565b60208301525092915050565b5f611778611773846116bc565b6116a2565b9050808382526020820190506040840283018581111561179b5761179a6116e7565b5b835b818110156117c457806117b08882611719565b84526020840193505060408101905061179d565b5050509392505050565b5f82601f8301126117e2576117e1611630565b5b81356117f2848260208601611766565b91505092915050565b5f602082840312156118105761180f6115d3565b5b5f82013567ffffffffffffffff81111561182d5761182c6115d7565b5b611839848285016117ce565b91505092915050565b5f81519050919050565b5f82825260208201905092915050565b5f819050602082019050919050565b61187481611591565b82525050565b6118838161153d565b82525050565b604082015f82015161189d5f85018261186b565b5060208201516118b0602085018261187a565b50505050565b5f6118c18383611889565b60408301905092915050565b5f602082019050919050565b5f6118e382611842565b6118ed818561184c565b93506118f88361185c565b805f5b8381101561192857815161190f88826118b6565b975061191a836118cd565b9250506001810190506118fb565b5085935050505092915050565b5f6020820190508181035f83015261194d81846118d9565b905092915050565b5f82825260208201905092915050565b7f6f6e6c79206f776e6572000000000000000000000000000000000000000000005f82015250565b5f611999600a83611955565b91506119a482611965565b602082019050919050565b5f6020820190508181035f8301526119c68161198d565b9050919050565b5f819050919050565b6119df816119cd565b81146119e9575f5ffd5b50565b5f815190506119fa816119d6565b92915050565b5f60208284031215611a1557611a146115d3565b5b5f611a22848285016119ec565b91505092915050565b611a34816119cd565b82525050565b5f602082019050611a4d5f830184611a2b565b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601160045260245ffd5b5f611a8a826119cd565b9150611a95836119cd565b9250828202611aa3816119cd565b91508282048414831517611aba57611ab9611a53565b5b5092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52601260045260245ffd5b5f611af8826119cd565b9150611b03836119cd565b925082611b1357611b12611ac1565b5b828204905092915050565b5f611b28826119cd565b9150611b33836119cd565b9250828203905081811115611b4b57611b4a611a53565b5b92915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52603260045260245ffd5b5f611b88826119cd565b9150611b93836119cd565b9250828201905080821115611bab57611baa611a53565b5b92915050565b5f604082019050611bc45f8301856115a2565b611bd16020830184611a2b565b9392505050565b5f81519050919050565b5f81905092915050565b8281835e5f83830152505050565b5f611c0482611bd8565b611c0e8185611be2565b9350611c1e818560208601611bec565b80840191505092915050565b5f611c358284611bfa565b915081905092915050565b5f8115159050919050565b611c5481611c40565b8114611c5e575f5ffd5b50565b5f81519050611c6f81611c4b565b92915050565b5f60208284031215611c8a57611c896115d3565b5b5f611c9784828501611c61565b91505092915050565b5f819050919050565b5f611cc3611cbe611cb98461153d565b611ca0565b6119cd565b9050919050565b611cd381611ca9565b82525050565b5f602082019050611cec5f830184611cca565b9291505056fea2646970667358221220db0e906ba1767c17ab00ccfd3244571f0eac240dcdde206c594dec6994dd655064736f6c634300081c0033",
}

// USDTRewardDistributorABI is the input ABI used to generate the binding from.
// Deprecated: Use USDTRewardDistributorMetaData.ABI instead.
var USDTRewardDistributorABI = USDTRewardDistributorMetaData.ABI

// USDTRewardDistributorBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use USDTRewardDistributorMetaData.Bin instead.
var USDTRewardDistributorBin = USDTRewardDistributorMetaData.Bin

// DeployUSDTRewardDistributor deploys a new Ethereum contract, binding an instance of USDTRewardDistributor to it.
func DeployUSDTRewardDistributor(auth *bind.TransactOpts, backend bind.ContractBackend, initialUSDT common.Address, initialFeeBps *big.Int, initialOwner common.Address, initialFeeWallet common.Address, initialRecipients []USDTRewardDistributorRecipient) (common.Address, *types.Transaction, *USDTRewardDistributor, error) {
	parsed, err := USDTRewardDistributorMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(USDTRewardDistributorBin), backend, initialUSDT, initialFeeBps, initialOwner, initialFeeWallet, initialRecipients)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &USDTRewardDistributor{USDTRewardDistributorCaller: USDTRewardDistributorCaller{contract: contract}, USDTRewardDistributorTransactor: USDTRewardDistributorTransactor{contract: contract}, USDTRewardDistributorFilterer: USDTRewardDistributorFilterer{contract: contract}}, nil
}

// USDTRewardDistributor is an auto generated Go binding around an Ethereum contract.
type USDTRewardDistributor struct {
	USDTRewardDistributorCaller     // Read-only binding to the contract
	USDTRewardDistributorTransactor // Write-only binding to the contract
	USDTRewardDistributorFilterer   // Log filterer for contract events
}

// USDTRewardDistributorCaller is an auto generated read-only Go binding around an Ethereum contract.
type USDTRewardDistributorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// USDTRewardDistributorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type USDTRewardDistributorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// USDTRewardDistributorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type USDTRewardDistributorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// USDTRewardDistributorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type USDTRewardDistributorSession struct {
	Contract     *USDTRewardDistributor // Generic contract binding to set the session for
	CallOpts     bind.CallOpts          // Call options to use throughout this session
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// USDTRewardDistributorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type USDTRewardDistributorCallerSession struct {
	Contract *USDTRewardDistributorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                // Call options to use throughout this session
}

// USDTRewardDistributorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type USDTRewardDistributorTransactorSession struct {
	Contract     *USDTRewardDistributorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                // Transaction auth options to use throughout this session
}

// USDTRewardDistributorRaw is an auto generated low-level Go binding around an Ethereum contract.
type USDTRewardDistributorRaw struct {
	Contract *USDTRewardDistributor // Generic contract binding to access the raw methods on
}

// USDTRewardDistributorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type USDTRewardDistributorCallerRaw struct {
	Contract *USDTRewardDistributorCaller // Generic read-only contract binding to access the raw methods on
}

// USDTRewardDistributorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type USDTRewardDistributorTransactorRaw struct {
	Contract *USDTRewardDistributorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewUSDTRewardDistributor creates a new instance of USDTRewardDistributor, bound to a specific deployed contract.
func NewUSDTRewardDistributor(address common.Address, backend bind.ContractBackend) (*USDTRewardDistributor, error) {
	contract, err := bindUSDTRewardDistributor(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &USDTRewardDistributor{USDTRewardDistributorCaller: USDTRewardDistributorCaller{contract: contract}, USDTRewardDistributorTransactor: USDTRewardDistributorTransactor{contract: contract}, USDTRewardDistributorFilterer: USDTRewardDistributorFilterer{contract: contract}}, nil
}

// NewUSDTRewardDistributorCaller creates a new read-only instance of USDTRewardDistributor, bound to a specific deployed contract.
func NewUSDTRewardDistributorCaller(address common.Address, caller bind.ContractCaller) (*USDTRewardDistributorCaller, error) {
	contract, err := bindUSDTRewardDistributor(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &USDTRewardDistributorCaller{contract: contract}, nil
}

// NewUSDTRewardDistributorTransactor creates a new write-only instance of USDTRewardDistributor, bound to a specific deployed contract.
func NewUSDTRewardDistributorTransactor(address common.Address, transactor bind.ContractTransactor) (*USDTRewardDistributorTransactor, error) {
	contract, err := bindUSDTRewardDistributor(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &USDTRewardDistributorTransactor{contract: contract}, nil
}

// NewUSDTRewardDistributorFilterer creates a new log filterer instance of USDTRewardDistributor, bound to a specific deployed contract.
func NewUSDTRewardDistributorFilterer(address common.Address, filterer bind.ContractFilterer) (*USDTRewardDistributorFilterer, error) {
	contract, err := bindUSDTRewardDistributor(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &USDTRewardDistributorFilterer{contract: contract}, nil
}

// bindUSDTRewardDistributor binds a generic wrapper to an already deployed contract.
func bindUSDTRewardDistributor(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := USDTRewardDistributorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_USDTRewardDistributor *USDTRewardDistributorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _USDTRewardDistributor.Contract.USDTRewardDistributorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_USDTRewardDistributor *USDTRewardDistributorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.USDTRewardDistributorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_USDTRewardDistributor *USDTRewardDistributorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.USDTRewardDistributorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_USDTRewardDistributor *USDTRewardDistributorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _USDTRewardDistributor.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_USDTRewardDistributor *USDTRewardDistributorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_USDTRewardDistributor *USDTRewardDistributorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.contract.Transact(opts, method, params...)
}

// FeeBps is a free data retrieval call binding the contract method 0x24a9d853.
//
// Solidity: function feeBps() view returns(uint16)
func (_USDTRewardDistributor *USDTRewardDistributorCaller) FeeBps(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _USDTRewardDistributor.contract.Call(opts, &out, "feeBps")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// FeeBps is a free data retrieval call binding the contract method 0x24a9d853.
//
// Solidity: function feeBps() view returns(uint16)
func (_USDTRewardDistributor *USDTRewardDistributorSession) FeeBps() (uint16, error) {
	return _USDTRewardDistributor.Contract.FeeBps(&_USDTRewardDistributor.CallOpts)
}

// FeeBps is a free data retrieval call binding the contract method 0x24a9d853.
//
// Solidity: function feeBps() view returns(uint16)
func (_USDTRewardDistributor *USDTRewardDistributorCallerSession) FeeBps() (uint16, error) {
	return _USDTRewardDistributor.Contract.FeeBps(&_USDTRewardDistributor.CallOpts)
}

// FeeWallet is a free data retrieval call binding the contract method 0xf25f4b56.
//
// Solidity: function feeWallet() view returns(address)
func (_USDTRewardDistributor *USDTRewardDistributorCaller) FeeWallet(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _USDTRewardDistributor.contract.Call(opts, &out, "feeWallet")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// FeeWallet is a free data retrieval call binding the contract method 0xf25f4b56.
//
// Solidity: function feeWallet() view returns(address)
func (_USDTRewardDistributor *USDTRewardDistributorSession) FeeWallet() (common.Address, error) {
	return _USDTRewardDistributor.Contract.FeeWallet(&_USDTRewardDistributor.CallOpts)
}

// FeeWallet is a free data retrieval call binding the contract method 0xf25f4b56.
//
// Solidity: function feeWallet() view returns(address)
func (_USDTRewardDistributor *USDTRewardDistributorCallerSession) FeeWallet() (common.Address, error) {
	return _USDTRewardDistributor.Contract.FeeWallet(&_USDTRewardDistributor.CallOpts)
}

// GetRecipients is a free data retrieval call binding the contract method 0xd78d610b.
//
// Solidity: function getRecipients() view returns((address,uint16)[])
func (_USDTRewardDistributor *USDTRewardDistributorCaller) GetRecipients(opts *bind.CallOpts) ([]USDTRewardDistributorRecipient, error) {
	var out []interface{}
	err := _USDTRewardDistributor.contract.Call(opts, &out, "getRecipients")

	if err != nil {
		return *new([]USDTRewardDistributorRecipient), err
	}

	out0 := *abi.ConvertType(out[0], new([]USDTRewardDistributorRecipient)).(*[]USDTRewardDistributorRecipient)

	return out0, err

}

// GetRecipients is a free data retrieval call binding the contract method 0xd78d610b.
//
// Solidity: function getRecipients() view returns((address,uint16)[])
func (_USDTRewardDistributor *USDTRewardDistributorSession) GetRecipients() ([]USDTRewardDistributorRecipient, error) {
	return _USDTRewardDistributor.Contract.GetRecipients(&_USDTRewardDistributor.CallOpts)
}

// GetRecipients is a free data retrieval call binding the contract method 0xd78d610b.
//
// Solidity: function getRecipients() view returns((address,uint16)[])
func (_USDTRewardDistributor *USDTRewardDistributorCallerSession) GetRecipients() ([]USDTRewardDistributorRecipient, error) {
	return _USDTRewardDistributor.Contract.GetRecipients(&_USDTRewardDistributor.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_USDTRewardDistributor *USDTRewardDistributorCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _USDTRewardDistributor.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_USDTRewardDistributor *USDTRewardDistributorSession) Owner() (common.Address, error) {
	return _USDTRewardDistributor.Contract.Owner(&_USDTRewardDistributor.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_USDTRewardDistributor *USDTRewardDistributorCallerSession) Owner() (common.Address, error) {
	return _USDTRewardDistributor.Contract.Owner(&_USDTRewardDistributor.CallOpts)
}

// Usdt is a free data retrieval call binding the contract method 0x2f48ab7d.
//
// Solidity: function usdt() view returns(address)
func (_USDTRewardDistributor *USDTRewardDistributorCaller) Usdt(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _USDTRewardDistributor.contract.Call(opts, &out, "usdt")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Usdt is a free data retrieval call binding the contract method 0x2f48ab7d.
//
// Solidity: function usdt() view returns(address)
func (_USDTRewardDistributor *USDTRewardDistributorSession) Usdt() (common.Address, error) {
	return _USDTRewardDistributor.Contract.Usdt(&_USDTRewardDistributor.CallOpts)
}

// Usdt is a free data retrieval call binding the contract method 0x2f48ab7d.
//
// Solidity: function usdt() view returns(address)
func (_USDTRewardDistributor *USDTRewardDistributorCallerSession) Usdt() (common.Address, error) {
	return _USDTRewardDistributor.Contract.Usdt(&_USDTRewardDistributor.CallOpts)
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns()
func (_USDTRewardDistributor *USDTRewardDistributorTransactor) Distribute(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _USDTRewardDistributor.contract.Transact(opts, "distribute")
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns()
func (_USDTRewardDistributor *USDTRewardDistributorSession) Distribute() (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.Distribute(&_USDTRewardDistributor.TransactOpts)
}

// Distribute is a paid mutator transaction binding the contract method 0xe4fc6b6d.
//
// Solidity: function distribute() returns()
func (_USDTRewardDistributor *USDTRewardDistributorTransactorSession) Distribute() (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.Distribute(&_USDTRewardDistributor.TransactOpts)
}

// ReplaceRecipients is a paid mutator transaction binding the contract method 0xa6e845d1.
//
// Solidity: function replaceRecipients((address,uint16)[] newRecipients) returns()
func (_USDTRewardDistributor *USDTRewardDistributorTransactor) ReplaceRecipients(opts *bind.TransactOpts, newRecipients []USDTRewardDistributorRecipient) (*types.Transaction, error) {
	return _USDTRewardDistributor.contract.Transact(opts, "replaceRecipients", newRecipients)
}

// ReplaceRecipients is a paid mutator transaction binding the contract method 0xa6e845d1.
//
// Solidity: function replaceRecipients((address,uint16)[] newRecipients) returns()
func (_USDTRewardDistributor *USDTRewardDistributorSession) ReplaceRecipients(newRecipients []USDTRewardDistributorRecipient) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.ReplaceRecipients(&_USDTRewardDistributor.TransactOpts, newRecipients)
}

// ReplaceRecipients is a paid mutator transaction binding the contract method 0xa6e845d1.
//
// Solidity: function replaceRecipients((address,uint16)[] newRecipients) returns()
func (_USDTRewardDistributor *USDTRewardDistributorTransactorSession) ReplaceRecipients(newRecipients []USDTRewardDistributorRecipient) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.ReplaceRecipients(&_USDTRewardDistributor.TransactOpts, newRecipients)
}

// SetFeeWallet is a paid mutator transaction binding the contract method 0x90d49b9d.
//
// Solidity: function setFeeWallet(address newFeeWallet) returns()
func (_USDTRewardDistributor *USDTRewardDistributorTransactor) SetFeeWallet(opts *bind.TransactOpts, newFeeWallet common.Address) (*types.Transaction, error) {
	return _USDTRewardDistributor.contract.Transact(opts, "setFeeWallet", newFeeWallet)
}

// SetFeeWallet is a paid mutator transaction binding the contract method 0x90d49b9d.
//
// Solidity: function setFeeWallet(address newFeeWallet) returns()
func (_USDTRewardDistributor *USDTRewardDistributorSession) SetFeeWallet(newFeeWallet common.Address) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.SetFeeWallet(&_USDTRewardDistributor.TransactOpts, newFeeWallet)
}

// SetFeeWallet is a paid mutator transaction binding the contract method 0x90d49b9d.
//
// Solidity: function setFeeWallet(address newFeeWallet) returns()
func (_USDTRewardDistributor *USDTRewardDistributorTransactorSession) SetFeeWallet(newFeeWallet common.Address) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.SetFeeWallet(&_USDTRewardDistributor.TransactOpts, newFeeWallet)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_USDTRewardDistributor *USDTRewardDistributorTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _USDTRewardDistributor.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_USDTRewardDistributor *USDTRewardDistributorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.TransferOwnership(&_USDTRewardDistributor.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_USDTRewardDistributor *USDTRewardDistributorTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.TransferOwnership(&_USDTRewardDistributor.TransactOpts, newOwner)
}

// WithdrawUSDT is a paid mutator transaction binding the contract method 0x59db80f8.
//
// Solidity: function withdrawUSDT(address destination) returns()
func (_USDTRewardDistributor *USDTRewardDistributorTransactor) WithdrawUSDT(opts *bind.TransactOpts, destination common.Address) (*types.Transaction, error) {
	return _USDTRewardDistributor.contract.Transact(opts, "withdrawUSDT", destination)
}

// WithdrawUSDT is a paid mutator transaction binding the contract method 0x59db80f8.
//
// Solidity: function withdrawUSDT(address destination) returns()
func (_USDTRewardDistributor *USDTRewardDistributorSession) WithdrawUSDT(destination common.Address) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.WithdrawUSDT(&_USDTRewardDistributor.TransactOpts, destination)
}

// WithdrawUSDT is a paid mutator transaction binding the contract method 0x59db80f8.
//
// Solidity: function withdrawUSDT(address destination) returns()
func (_USDTRewardDistributor *USDTRewardDistributorTransactorSession) WithdrawUSDT(destination common.Address) (*types.Transaction, error) {
	return _USDTRewardDistributor.Contract.WithdrawUSDT(&_USDTRewardDistributor.TransactOpts, destination)
}
