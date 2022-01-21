## Running integration tests

### Prerequisites
You need Node v14+ and yarn installed on your machine.  
You also need to install the project's dependencies via `yarn install`.

### Starting integration tests
Run `make test-integration`.  
This script will spin a new blockchain on your local machine and run all integration tests.

### Run specific test

`./scripts/test/run-test-integration.sh -t '<description name> <it-name>`  
(eg. `./scripts/test/run-test-integration.sh -t 'native transfers validator can send tokens'`)

Same as above, but the script will pass the extra -t argument to `jest`, which tells it to only run the specified test.

## Add proto file codec

In order to send custom proto queries/transactions, you have to reference their respective proto files in the `./scripts/define-custom-proto.sh`  script. Then run the job `npm run setup-proto`, which generates proto codes in `src/util/codec`. Finally, add the codecs to the custom registry in `clients.ts`. 