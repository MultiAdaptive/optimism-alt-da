# How to Use the OP Stack with multiAdaptive

## Software Dependencies  
| Dependency | Version | Version Check Command |
|------------|---------|-----------------------|
| git        | ^2      | git --version         |
| go	        | ^1.21	  | go version            |
| node       | ^20	    | node --version        |
| pnpm       | ^8	     | pnpm --version        |
| foundry    | ^0.2.0  | forge --version       |
| make       | ^3      | make --version        |
| jq         | ^1.6    | jq --version          |
| direnv     | ^2      | direnv --version      |

## Compile the Core Codebase  
Setting up the EVM Rollup requires compiling code from two critical repositories: the [optimism-Alt-DA](https://github.com/MultiAdaptive/optimism-alt-da) monorepo and the [op-geth](https://github.com/ethereum-optimism/op-geth) repository.  

## Build the Adapter Source

1. Clone the Optimism Monorepo  
    ```Base
    cd ~
    git clone https://github.com/MultiAdaptive/optimism-alt-da.git
    cd optimism-alt-da
    ```
2. Install modules
    ```Base
   pnpm install
    ```
3. Compile the necessary packages 
    ```Base
    make op-node op-batcher op-proposer
    pnpm build
    ```

## Build the Optimism Geth Source
1. Clone and navigate to op-geth  
    ```Base
    cd ~
    git clone https://github.com/ethereum-optimism/op-geth.git
    cd op-geth
    ```
2. Compile op-geth
    ```Base
    make geth
    ```

## Fill Out Environment Variables  
1. Enter the Optimism Monorepo
    ```Base
    cd ~/optimism-alt-da
    ```
2. Duplicate the sample environment variable file  
    ```Base
    cp .envrc.example .envrc
    ```  
3. Fill out the environment variable file  
   Open up the environment variable file and fill out the following variables: 

| Variable Name | Description                                                                                                                                                                                |
|---------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| L1_RPC_URL    | 	URL for your L1 node (a Sepolia node in this case).                                                                                                                                       |
| L1_RPC_KIND   | Kind of L1 RPC you're connecting to, used to inform optimal transactions receipts fetching. Valid options: alchemy, quicknode, infura, parity, nethermind, debug_geth, erigon, basic, any. |


## Generate and Secure Keys  
Create four essential accounts with private keys:  
* The Admin address has the ability to upgrade contracts.  
* The Batcher address publishes Sequencer transaction data to L1.  
* The Proposer address publishes L2 transaction results (state roots) to L1.  
* The Sequencer address signs blocks on the p2p network.

1. Enter the Optimism Monorepo
    ```Base
    cd ~/optimism-alt-da
    ```
2. Generate new addresses
    ```Base
    ./packages/contracts-bedrock/scripts/getting-started/wallets.sh
    ```  
3. Check the output  
   Make sure that you see output that looks something like the following:  
   ```Base
   # Copy the following into your .envrc file:
   
   # Admin account
   export GS_ADMIN_ADDRESS=0x798B181B03831eBB13C417FC53c58329D4bBfc76
   export GS_ADMIN_PRIVATE_KEY=0x2edd47dc20882a112025c1098742b6fec34b3d93140bfb454a6de293ca260730
   
   # Batcher account
   export GS_BATCHER_ADDRESS=0x14d6499Ad02613023Ebce461043833D15c78c645
   export GS_BATCHER_PRIVATE_KEY=0x2b9edc56f44992850b9e3039d388ec6c44568fc28f503e129a9a5b620e1c6358
   
   # Proposer account
   export GS_PROPOSER_ADDRESS=0xEc62924245D9B42A10138f9146354001e1410847
   export GS_PROPOSER_PRIVATE_KEY=0xec7ef5c45db1510b3784b4d91f1798dafb4ed8d957a7930c027e325ec9432f81
   
   # Sequencer account
   export GS_SEQUENCER_ADDRESS=0x23788522aA1A460a584D34fe3ad55b0b0Bc033e8
   export GS_SEQUENCER_PRIVATE_KEY=0x01b8fa0c75ea22e5da21d3bd060c9b16736d8af9f18d36cea6677b76aad7c38e
   
   ```
4. Save the addresses  
   Copy the output from the previous step and paste it into your .envrc file as directed.
5. Fund the addresses  
   You will need to send ETH to the Admin, Proposer, and Batcher addresses. The exact amount of ETH required depends on the L1 network being used. You do not need to send any ETH to the Sequencer address as it does not send transactions.  
   It's recommended to fund the addresses with the following amounts when using Sepolia:  
   * Admin — 0.5 Sepolia ETH
   * Proposer — 0.2 Sepolia ETH
   * Batcher — 0.1 Sepolia ETH

## Load Environment variables
1. Enter the Optimism Monorepo  
    ```Base
    cd ~/optimism-alt-da
    ```
2. Load the variables with direnv  
    ```Base
    direnv allow
    ```
## Core Contract Deployment  
Deploy essential L1 contracts for the chain’s functionality:  
1. Update ~/optimism-alt-da/packages/contracts-bedrock/deploy-config and update file multiadaptive.json by referencing the following configs ( addresses will vary for each user, the following is a demo-config deployed for sepolia L1)  
   ```json
   {
     "l1StartingBlockTag": "0x32e7e9c4320581f9a097cb2d03f420219b6f2793b9929fd24e0cbe437cdd7a95",
   
     "l1ChainID": 11155111,
     "l2ChainID": 1198,
     "l2BlockTime": 2,
     "l1BlockTime": 12,
   
     "maxSequencerDrift": 600,
     "sequencerWindowSize": 3600,
     "channelTimeout": 300,
     
     "p2pSequencerAddress": "0x20242D81D7F32FD1D86DA5fAC5D6Bba021B58a9f",
     "batchInboxAddress": "0xff00000000000000000000000000000011155420",
     "batchSenderAddress": "0xE14b3f075AD9377689dAf659e04A2a99a4AcEde4",
     
     "l2OutputOracleSubmissionInterval": 120,
     "l2OutputOracleStartingBlockNumber": 0,
     "l2OutputOracleStartingTimestamp": 1723695456,
   
     "l2OutputOracleProposer": "0xEc4C0983dddBf6FED78385e3ee2d767d12377342",
     "l2OutputOracleChallenger": "0xE4Ff7382cc4721364540c9E5fF2a86201c0B0363",
   
     "finalizationPeriodSeconds": 12,
   
   
     "proxyAdminOwner": "0xE4Ff7382cc4721364540c9E5fF2a86201c0B0363",
     "baseFeeVaultRecipient": "0xE4Ff7382cc4721364540c9E5fF2a86201c0B0363",
     "l1FeeVaultRecipient": "0xE4Ff7382cc4721364540c9E5fF2a86201c0B0363",
     "sequencerFeeVaultRecipient": "0xE4Ff7382cc4721364540c9E5fF2a86201c0B0363",
     "finalSystemOwner": "0xE4Ff7382cc4721364540c9E5fF2a86201c0B0363",
     "superchainConfigGuardian": "0xE4Ff7382cc4721364540c9E5fF2a86201c0B0363",
   
     "baseFeeVaultMinimumWithdrawalAmount": "0x8ac7230489e80000",
     "l1FeeVaultMinimumWithdrawalAmount": "0x8ac7230489e80000",
     "sequencerFeeVaultMinimumWithdrawalAmount": "0x8ac7230489e80000",
     "baseFeeVaultWithdrawalNetwork": 0,
     "l1FeeVaultWithdrawalNetwork": 0,
     "sequencerFeeVaultWithdrawalNetwork": 0,
   
     "enableGovernance": true,
     "governanceTokenSymbol": "OP",
     "governanceTokenName": "Optimism",
     "governanceTokenOwner": "0xE4Ff7382cc4721364540c9E5fF2a86201c0B0363",
   
     "l2GenesisBlockGasLimit": "0x1c9c380",
     "l2GenesisBlockBaseFeePerGas": "0x3b9aca00",
   
     "eip1559Denominator": 50,
     "eip1559Elasticity": 6,
   
     "l2GenesisRegolithTimeOffset": "0x0",
     "systemConfigStartBlock": 0,
   
     "requiredProtocolVersion": "0x0000000000000000000000000000000000000004000000000000000000000001",
     "recommendedProtocolVersion": "0x0000000000000000000000000000000000000004000000000000000000000001",
   
     "fundDevAccounts": true,
     "faultGameAbsolutePrestate": "0x03c7ae758795765c6664a5d39bf63841c71ff191e9189522bad8ebff5d4eca98",
     "faultGameMaxDepth": 44,
     "faultGameClockExtension": 0,
     "faultGameMaxClockDuration": 600,
     "faultGameGenesisBlock": 0,
     "faultGameGenesisOutputRoot": "0x0000000000000000000000000000000000000000000000000000000000000000",
     "faultGameSplitDepth": 14,
     "faultGameWithdrawalDelay": 604800,
   
     "preimageOracleMinProposalSize": 1800000,
     "preimageOracleChallengePeriod": 86400,
   
     "proofMaturityDelaySeconds": 604800,
     "disputeGameFinalityDelaySeconds": 302400,
     "respectedGameType": 0,
     "useFaultProofs": true,
     "usePlasma": true,
     "daCommitmentType": "GenericCommitment",
     "daChallengeWindow": 160,
     "daResolveWindow": 160,
     "daBondSize": 1000000,
     "daResolverRefundPercentage": 0,
     "cliqueSignerAddress": "0xE4Ff7382cc4721364540c9E5fF2a86201c0B0363",
     "l1UseClique": true,
     "gasPriceOracleOverhead": 2100,
     "gasPriceOracleScalar": 1000000,
     "gasPriceOracleBaseFeeScalar": 1368,
     "gasPriceOracleBlobBaseFeeScalar": 810949,
     "eip1559DenominatorCanyon": 250,
     "l2GenesisDeltaTimeOffset": "0x0",
     "l2GenesisCanyonTimeOffset": "0x0",
     "l2GenesisEcotoneTimeOffset": "0x0"
   }
   
   ```
2. Navigate to ~/optimism-alt-da/packages/contracts-bedrock/ and the deploy contracts (this can take up to 15 minutes)  
   ```Base
      cd ~/optimism-alt-da/packages/contracts-bedrock/
      DEPLOYMENT_OUTFILE=deployments/artifact.json \
      DEPLOY_CONFIG_PATH=deploy-config/multiadaptive.json \
      forge script scripts/Deploy.s.sol:Deploy  --broadcast --private-key $GS_ADMIN_PRIVATE_KEY \
      --rpc-url $L1_RPC_URL --slow
   ```  
3. L2 Allocs 
   ```Base
   CONTRACT_ADDRESSES_PATH=deployments/artifact.json \
   DEPLOY_CONFIG_PATH=deploy-config/multiadaptive.json \
   STATE_DUMP_PATH=deploy-config/l2-allocs.json \
   forge script scripts/L2Genesis.s.sol:L2Genesis \
   --sig 'runWithStateDump()' --chain-id $L2_CHAIN_ID
   ```

## Setting Up L2 Configuration  
Now that you've set up the L1 smart contracts you can automatically generate several configuration files that are used within the Consensus Client and the Execution Client.  
You need to generate three important files:  
* genesis.json includes the genesis state of the chain for the Execution Client.  
* rollup.json includes configuration information for the Consensus Client.
* jwt.txt is a [JSON Web Token](https://jwt.io/introduction) that allows the Consensus Client and the Execution Client to communicate securely (the same mechanism is used in Ethereum clients).

1. Create genesis files   
   ```Base
   ../../op-node/bin/op-node genesis l2 \
   --l1-rpc $L1_RPC_URL \
   --deploy-config deploy-config/multiadaptive.json \
   --l2-allocs deploy-config/l2-allocs.json \
   --l1-deployments deployments/artifact.json \
   --outfile.l2 ../../op-node/genesis.json \
   --outfile.rollup ../../op-node/rollup.json
   ```
2. Create an authentication key  
   Next you'll create a [JSON Web Token](https://jwt.io/introduction) that will be used to authenticate the Consensus Client and the Execution Client. This token is used to ensure that only the Consensus Client and the Execution Client can communicate with each other. You can generate a JWT with the following command:
   ```Base
   openssl rand -hex 32 > jwt.txt
   ```
   
3. Copy genesis files into the op-geth directory  
   Finally, you'll need to copy the genesis.json file and jwt.txt file into op-geth so you can use it to initialize and run op-geth:
   ```Base
   cp genesis.json ~/op-geth
   cp jwt.txt ~/op-geth
   ```

# Initialize op-geth
You're almost ready to run your chain! Now you just need to run a few commands to initialize op-geth. You're going to be running a Sequencer node, so you'll need to import the Sequencer private key that you generated earlier. This private key is what your Sequencer will use to sign new blocks.  
1. Navigate to the op-geth directory
   ```Base
   cd ~/op-geth
   ```
2. Create a data directory folder
   ```Base
   mkdir datadir
   ```
3. Initialize op-geth
   ```Base
   build/bin/geth init --datadir=datadir genesis.json
   ```
   
## Start MultiAdaptive TransitStation
MultiAdaptive TransitStation will help you to submit your DA data to MultiAdaptive.

Please refer to the [documentation](https://github.com/MultiAdaptive/transitStation/blob/main/README.md) for details.

## Start op-geth
Now you'll start op-geth, your Execution Client. Note that you won't start seeing any transactions until you start the Consensus Client in the next step.
1. Open up a new terminal 
   You'll need a terminal window to run op-geth in.
2. Navigate to the op-geth directory
   ```Bsse
   cd ~/op-geth
   ```
3. Run op-geth  
   ```Base
   ./build/bin/geth \                                
   --datadir ./datadir \
   --http \
   --http.corsdomain="*" \
   --http.vhosts="*" \
   --http.addr=0.0.0.0 \
   --http.api=web3,debug,eth,txpool,net,engine \
   --ws \
   --ws.addr=0.0.0.0 \
   --ws.port=8546 \
   --ws.origins="*" \
   --ws.api=debug,eth,txpool,net,engine \
   --syncmode=full \
   --gcmode=archive \
   --nodiscover \
   --maxpeers=0 \
   --networkid=42069 \
   --authrpc.vhosts="*" \
   --authrpc.addr=0.0.0.0 \
   --authrpc.port=8551 \
   --authrpc.jwtsecret=./jwt.txt \
   --rollup.disabletxpoolgossip=true
   ```
## Start op-node  
1. Open up a new terminal
   You'll need a terminal window to run op-geth in.
2. Navigate to the op-node directory
   ```Bsse
   cd ~/optimism/op-node
   ```
3. Run op-node
   ```Base
   ./bin/op-node \
   --l2=http://localhost:8551 \
   --l2.jwt-secret=./jwt.txt \
   --sequencer.enabled \
   --sequencer.l1-confs=5 \
   --verifier.l1-confs=4 \
   --rollup.config=./rollup.json \
   --rpc.addr=0.0.0.0 \
   --p2p.disable \
   --rpc.enable-admin \
   --p2p.sequencer.key=$GS_SEQUENCER_PRIVATE_KEY \
   --l1=$L1_RPC_URL \
   --l1.rpckind=$L1_RPC_KIND \
   --l1.beacon=<L1_Beacan_URL> \
   --plasma.enabled=true \
   --plasma.double-send=true \
   --plasma.da-server=<TransitStation_SERVER_URL>
   ```

## Start op-batcher
1. Open up a new terminal
   You'll need a terminal window to run op-batcher in.
2. Navigate to the op-batcher directory
   ```Bsse
   cd ~/optimism/op-batcher
   ```
3. Run op-node
   ```Base
   ./bin/op-batcher \
   --l2-eth-rpc=http://localhost:8545 \
   --rollup-rpc=http://localhost:9545 \
   --poll-interval=1s \
   --sub-safety-margin=6 \
   --num-confirmations=1 \
   --safe-abort-nonce-too-low-count=3 \
   --resubmission-timeout=30s \
   --rpc.addr=0.0.0.0 \
   --rpc.port=8548 \
   --rpc.enable-admin \
   --max-channel-duration=25 \
   --l1-eth-rpc=$L1_RPC_URL \
   --private-key=$GS_BATCHER_PRIVATE_KEY \
   --data-availability-type=blobs \
   --plasma.enabled=true \
   --plasma.da-service=true \
   --plasma.double-send=true \
   --plasma.da-server=<TransitStation_SERVER_URL>
   ```
   
## Start op-proposer
1. Open up a new terminal
   You'll need a terminal window to run op-batcher in.
2. Navigate to the op-proposer directory
   ```Bsse
   cd ~/optimism/op-proposer
   ```
3. Run op-proposer
   ```Base
   ./bin/op-proposer \
   --poll-interval=12s \
   --rpc.port=8560 \
   --rollup-rpc=http://localhost:9545 \
   --l2oo-address=$(cat ../packages/contracts-bedrock/deployments/artifact.json | jq -r .L2OutputOracleProxy) \
   --private-key=$GS_PROPOSER_PRIVATE_KEY \
   --l1-eth-rpc=$L1_RPC_URL
   ```
