
## Monitor Service
this code will be the model so we can create a smart contract monitor that will check out all the transactions that happen to certain contract address that we can set at the file's top level. It will have a list of endpoints it can check from, while prioritizing certain endpoint, it can try to query again from another if it fails twice and proceed rotating. It queries the contract's last transactions every 5 seconds and whenever a transaction with the following action name happens, it will update a planetscale mysql database by adding the transaction to storage - transaction hash, an array of asset ids, timestamp, account that has sent, action name, and processed true or false

### Stake Create and Remove Flow
when the action name is stake create, the flow is as follows: put the transaction to storage with processed as false, add each of the asset ids as a table row with their corresponding account sender name as owner and whatever extra data necessary. after that set processed to true.
when stakeremove is called: put the transaction to storage with processed as false,  erase the asset from the table, after that set processed to true.

## Rewards Service

This rewards service operates a cron job that runs every hour to distribute Breeders tokens to NFT stakers.

It's logic is based on two components:
    Transaction pushing
    Reliability

### Transaction Pushing

### Reliability

#### Retry Mechanism
I've implemented a system that will attempt to send the transaction 3 times before giving up. The intervals between each try are 10 seconds, 20 seconds and 30 seconds. There's no issue if the transaction somehow takes longer to be approved, as the smart contract itself will not allow it to be pushed twice to the chain in such short time - minimum period is 1 hour.

#### Notification System and Logging
If the transaction fails to be pushed, the service will do an SMTP call to the configured email address, notifying the failure to the admin. This is a very important feature, as it allows the admin to be aware of any issues with the service in real time.

Every 10 minutes, the service will attempt to send some GET queries to the chain, and will generate notifications if the chain is not responding. 

All logs are stored locally in a file called logs.json and they are also sent to the admin's telegram account.

#### Limitations
As the service is fairly simple, I've opted not to monitor CPU and memory usage. The most important part of it is to allow the admin to be aware of any issues with the service. 
I am running the service locally at a VPS, without any sort of containerization. Graceful restarts are implemented, but the service will not be able to recover from a server crash.

## Tasklist
- [x] Fix main function
- [ ] Add db logic
    - [x] Add db connection
    - [x] Add db crud operations
    - [x] Add db schema
    - [x] Setup planetscale db
    - [ ] Setup automatic planetscale deployment for new instances
- [x] Add on-chain actions logic - even though what matters the most is the action.traces.data
- [ ] Add contract monitor logic
    - [x] Add contract monitor logic
    - [ ] Add contract monitor logic tests
    - [ ] Add contract monitor logic documentation
- [x] Add email notification service
    - [ ] Add smarter authentication through refresh tokens
    - [ ] Integrate service with actions and monitor
- [ ] Add telegram notification service
- [ ] Add easy to use config file
- [ ] Make it a CLI tool
- [ ] Create a flow to easily deploy new contract monitor instances
- [ ] Create an ABI scraper that will automatically find the data types of the contract's actions and tables, and create a model for it
