## PENDING

BREAKING CHANGES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [gaia-lite] [\#2182] Renamed and merged all redelegations endpoints into `/stake/redelegations`

* Gaia CLI  (`gaiacli`)

  * [\#810](https://github.com/cosmos/cosmos-sdk/issues/810) Don't fallback to any default values for chain ID.
    * Users need to supply chain ID either via config file or the `--chain-id` flag.
    * Change `chain_id` and `trust_node` in `gaiacli` configuration to `chain-id` and `trust-node` respectively.
  * [\#3069](https://github.com/cosmos/cosmos-sdk/pull/3069) `--fee` flag renamed to `--fees` to support multiple coins
  * [\#3156](https://github.com/cosmos/cosmos-sdk/pull/3156) Remove unimplemented `gaiacli init` command

* Gaia

* SDK

* Tendermint


FEATURES

* Gaia REST API (`gaiacli advanced rest-server`)
  * [\#3067](https://github.com/cosmos/cosmos-sdk/issues/3067) Add support for fees on transactions
  * [\#3069](https://github.com/cosmos/cosmos-sdk/pull/3069) Add a custom memo on transactions

* Gaia CLI  (`gaiacli`)
  * \#2399 Implement `params` command to query slashing parameters.

* Gaia

    * [\#2182] [x/stake] Added querier for querying a single redelegation

* SDK

* Tendermint


IMPROVEMENTS

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK

* Tendermint


BUG FIXES

* Gaia REST API (`gaiacli advanced rest-server`)

* Gaia CLI  (`gaiacli`)

* Gaia

* SDK

* Tendermint
