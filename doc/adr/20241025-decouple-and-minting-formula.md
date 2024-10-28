# Decouple from external libraries and upgrade minting formula

- Status: Accepted
- Deciders: Product Owner, Development Team
- Date: 2024-10-25
- Tags: minting formula, refactorings, upgrade

Technical story:
To improve the efficiency and sustainability of our platform, we are upgrading the minting formula, which will be effective starting from month 17 (https://www.wolframalpha.com/input?i=integral+-0.11175+x%5E3+%2B+50.82456+x%5E2+-+1767.49981+x+%2B+0.83381%C3%9710%5E6+dx+from+x+%3D+17+to+120). This upgrade aims to support a more balanced token distribution over time. Alongside this change, we are also decoupling from various external dependencies, such as WASM forks and Neutron modules, to increase control and facilitate future enhancements without third-party limitations.

These upgrades represent an ongoing initiative that has evolved over several months.

## Context and Problem Statement

Our decision to decouple from external libraries and wasm forks aligns with our objective of achieving greater control, adaptability, and security in the codebase. Integrating these dependencies into our core codebase allows us to fine-tune functionality to meet our specific needs and streamline updates.
The changes in the minting formula will allow for a more sustainable and fair distribution of tokens. 

By upgrading the minting formula, we aim to promote a fairer and more sustainable distribution of tokens. This formula prepares us for long-term growth.

Notable features and fixes provided by the upgrade include:

By internalizing neutron modules, we can customize aspects like unordered Interchain Accounts (ICA) channels and make direct modifications to IBC-related components as needed.

The minting formula upgrade supports an improved reward structure designed to encourage ecosystem stability and sustainability.

## Decision Drivers

Reduce dependency on external libraries and forks to increase code stability.

## Decision Outcome

The decision we took is to stop using forks of wasmd, neutron modules and upgrade the minting formula.

### Positive Consequences

- Greater Control: By internalizing dependencies, the development team has full access to and control over module updates and bug fixes, which allows for rapid issue resolution without waiting on external parties.

- Increased Reliability: With reduced external dependencies, we are less exposed to vulnerabilities or unexpected changes in external libraries.

### Negative Consequences

- Increased Maintenance Burden: Internalizing these components requires more responsibility from the development team to maintain and troubleshoot in-house code.

- Initial Refactoring Complexity: Migrating and refactoring external modules may introduce complexities and require rigorous testing to ensure smooth integration.

## Potential evolution

This decision marks a significant step toward platform independence and sustainability. Future enhancements may include, more decoupling, modules adjustments to the forked modules, documentation, and more.

This decision is documented to provide clarity on the rationale behind the upgrade and serves as a reference for future discussions and evaluations related to our blockchain development efforts.