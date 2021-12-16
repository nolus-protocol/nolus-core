workspace {

    model {
        user = person "User"
        nolusenterprise = enterprise "Nolus" {
            
            nolus = softwareSystem "Nolus" {
                validatornode = group "Validator/Sentry Node" {
                    cosmosapp = container "Cosmos App" {
                        bank = component "Bank"
                        tax_agent = component "Tax Agent"
                        minter = component "Minter"

                        tax_agent -> bank "distribute the transaction gas"
                        minter -> bank "mint amount on each block"
                    }
                    contracts = container "Smart Contracts" {
                        flex = component "Flex"
                        price_feed = component "Price Feed"
                        scheduler_data = component "Scheduler Data"
                        reserve_vault = component "Reserve Vault"
                        loans_vault = component "Loans Vault"
                    }
                    cosmosapp -> contracts "Execute Trx messages"
                    contracts -> cosmosapp "Store State"
                    contracts -> cosmosapp "Execute Trx messages"
                }
                appserver = container "Application Server" {
                    -> cosmosapp "Source Events"
                    -> cosmosapp "Forward Queries&Transactions"
                }

                webapp = container "Web UI Client" {
                    -> appserver "Queries, Transactions"
                }

                oracle_operator = container "Oracle Operator" {
                    market_data_aggregator = component "Market Data Aggregator"

                    -> appserver "Price updates"
                }
            }

            ibc = softwareSystem "Cosmos IBC Relay"

            ibc -> nolus "Relays-in"
            nolus -> ibc "Relays-out"

            admin = person "Admin" {
                -> webapp "Uses"
            }
        }

        user -> webapp "Uses"
    }

    views {
        systemContext nolus {
            include *
        }

        container nolus {
            include *
        }

        component cosmosapp {
            include *
        }

        component contracts {
            include *
        }

        component oracle_operator {
            include *
        }

        dynamic cosmosapp fee_handler {
            title "Tax & Inflation distribution"

            user -> tax_agent "send transaction"
            tax_agent -> reserve_vault "get Vault address"
            tax_agent -> bank "send extra fee to the Vault address"
            tax_agent -> bank "send remained gas to the Collector address"
            minter -> bank "[on block end] send newly minted coins to the Collector address"
        }

        dynamic cosmosapp oracle_msgs {
            title "Passing paid messages to oracle smart contracts"
            admin -> price_feed "update whitelist"

            market_data_aggregator -> price_feed "send observations"
            price_feed -> price_feed "match msg sender address to whitelist"
        }

        dynamic contracts "case0" "all" {
            title "Flex successful close"
            user -> flex "sign contract(amount, down-payment) && deposit down-pay"
            flex -> price_feed "get currency price"
            flex -> loans_vault "request loan"
            loans_vault -> flex "send amount/promise"
            user -> flex "repay one or more times until pay-off the total"
            flex -> user "transfer ownership"
            flex -> reserve_vault "send collateral"
            price_feed -> flex "push price update"
            scheduler_data -> flex "push end time period notification"
            autolayout
        }

        dynamic contracts "case1" "loan payment in a single epoch" {
            title "Loan payment in a single epoch"
            user -> flex "sign contract(amount, down-payment) && deposit down-pay"
            flex -> price_feed "get currency price"
            flex -> loans_vault "request loan"
            loans_vault -> flex "send amount/promise"
            user -> flex "repay one or more times until pay-off the total"
            flex -> user "transfer ownership"
            autolayout
        }

        dynamic contracts "case2" "update loans via oracles" {
            title "Update loans via oracles"
            price_feed -> flex "push price update"
            scheduler_data -> flex "push end time period notification"
            flex -> reserve_vault "send collateral"
            autolayout
        }
        theme default
    }
    
}