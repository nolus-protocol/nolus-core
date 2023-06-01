# Allocation

This module handles creating vesting accounts via the cli after the chain has started. The command differs from the cosmos-sdk's native one in that it allows for custom vesting start time.

This module is based on the [stargaze's alloc module](https://github.com/public-awesome/stargaze/tree/main/x/alloc).

## Governance Parameters

Here's a table for the parameters of the module:

| Name                       | Type                      | Description                                            | Default   |
| -------------------------- | ------------------------- | ------------------------------------------------------ | --------- |
