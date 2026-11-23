NETWORK ?= arbitrum
INPUT ?= input.json
CONTRACT ?=

# Read-only public defaults for the pinned fork tests. Override these with
# archive-capable endpoints when the defaults do not retain the required blocks.
ETHEREUM_RPC_URL ?= https://eth.drpc.org
ARBITRUM_RPC_URL ?= https://arbitrum-one.public.blastapi.io
export ETHEREUM_RPC_URL ARBITRUM_RPC_URL

.PHONY: gen test lint lint-go lint-solidity deploy set-distribution show distribute distribute-all verify

gen:
	forge build
	@tmp=$$(mktemp -d); trap 'rm -rf "$$tmp"' EXIT; \
	jq -c .abi out/USDTRewardDistributor.sol/USDTRewardDistributor.json > "$$tmp/distributor.abi"; \
	jq -r .bytecode.object out/USDTRewardDistributor.sol/USDTRewardDistributor.json > "$$tmp/distributor.bin"; \
	jq -c .abi out/USDTRewardDistributor.sol/IERC20Balance.json > "$$tmp/token.abi"; \
	mkdir -p internal/bindings; \
	abigen --abi "$$tmp/distributor.abi" --bin "$$tmp/distributor.bin" --pkg bindings --type USDTRewardDistributor --out internal/bindings/usdt_reward_distributor.go; \
	abigen --abi "$$tmp/token.abi" --pkg bindings --type IERC20Balance --out internal/bindings/ierc20_balance.go

pooler_args = --network "$(NETWORK)" --input "$(INPUT)" --contract "$(CONTRACT)"

deploy:
	@go run ./cmd/pooler deploy $(pooler_args)

set-distribution:
	@go run ./cmd/pooler set-distribution $(pooler_args)

show:
	@go run ./cmd/pooler show $(pooler_args)

distribute:
	@go run ./cmd/pooler distribute $(pooler_args)

distribute-all:
	@go run ./cmd/pooler distribute-all $(pooler_args)

verify:
	@go run ./cmd/pooler verify $(pooler_args)

lint: lint-go lint-solidity ## Runs linters

lint-go:
	go tool golangci-lint run ./...

lint-solidity:
	./node_modules/.bin/solhint --disc --noPoster 'contracts/**/*.sol'

test:
	@forge test
