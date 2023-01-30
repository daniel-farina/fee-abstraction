#!/usr/bin/env bash

osmosisd tx wasm store scripts/fee_abstraction.wasm --keyring-backend=test --home=$HOME/.osmosisd/validator1 --from deployer --chain-id testing --gas 10000000 --fees 25000stake --yes

sleep 5

ID=1

INIT='{"packet_lifetime":100}'
osmosisd tx wasm instantiate $ID "$INIT" --keyring-backend=test --home=$HOME/.osmosisd/validator1 --from deployer --chain-id testing --label "test" --no-admin --yes

sleep 5
CONTRACT=$(osmosisd query wasm list-contract-by-code $ID --output json | jq -r '.contracts[-1]')

query_params='{"query_stargate_twap":{"pool_id":1,"token_in_denom":"uosmo","token_out_denom":"uatom","with_swap_fee":false}}'
osmosisd query wasm contract-state smart $CONTRACT "$query_params"

echo "feeabs contract: "
echo $CONTRACT

sleep 2

if [[ $CONTRACT == null ]]
then 
    echo $CONTRACT
    echo "Contract deploy unsuccesful"
else
    echo wasm.$CONTRACT
    hermes --config scripts/relayer_hermes/config.toml create channel --a-chain testing --b-chain feeappd-t1 --a-port wasm.$CONTRACT --b-port feeabs --new-client-connection --yes
fi