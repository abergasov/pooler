# Read-only public defaults for the pinned fork tests. Override these with
# archive-capable endpoints when the defaults do not retain the required blocks.
ETHEREUM_RPC_URL ?= https://eth.drpc.org
ARBITRUM_RPC_URL ?= https://arbitrum-one.public.blastapi.io
export ETHEREUM_RPC_URL ARBITRUM_RPC_URL

.PHONY: test

test:
	@forge test
