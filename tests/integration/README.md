## Running integration tests

### Prerequisites
You need Node v14+ and yarn installed on your machine.  
You also need to install the project's dependencies via `yarn install`.

### Starting integration tests
Run `./run-integration.sh`.  
This script will spin a new blockchain on your local machine and run all integration tests.

### Run specific test
`./run-integration.sh -t '<description name> <it-name>`  
(eg. `./run-integration.sh -t 'native transfers validator can send tokens'`)

Same as above, but the script will pass the extra -t argument to `jest`, which tells it to only run the specified test.