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
                        timer = component "Timer"
                        reserve_vault = component "Reserve Vault"
                        loans_vault = component "Loans Vault"
                    }
                    cosmosapp -> contracts "Execute Trx messages"
                    contracts -> cosmosapp "Store State"
                    contracts -> cosmosapp "Execute Trx messages"
                }
                appserver = container "Application Server" {
                    api_endpoint = component "API Endpoint"

                    -> cosmosapp "Source Events"
                    -> cosmosapp "Forward Queries&Transactions"
                }

                webapp = container "Web UI Client" {
                    -> appserver "Queries, Transactions"
                }

                oracle_operator = container "Oracle Operator" {
                    market_data_operator = component "Market Data Operator" {
                       -> api_endpoint "Price updates"
                    }
                }
            }

            ibc = softwareSystem "Cosmos IBC Relay"

            ibc -> nolus "Relays-in"
            nolus -> ibc "Relays-out"

            admin = person "Admin" {
                -> webapp "Uses"
            }
        }

        market_data_aggregator = softwareSystem "Market Data Aggregator" {
        }

        market_data_operator -> market_data_aggregator "Fetch Data"

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

        dynamic contracts oracle_msgs {
            title "Price Feeds"

            admin -> price_feed "manage supported price pairs"
            admin -> price_feed "manage whitelisted operators"

            market_data_operator -> market_data_aggregator "poll observations"
            market_data_operator -> price_feed "send observations"
            price_feed -> price_feed "match msg sender address to whitelist"
            price_feed -> price_feed "update a price pair when aggregated observations pass % but not later than a delta t"
            price_feed -> flex "push price update"
        }

        dynamic contracts "case0" "all" {
            title "Flex successful close"
            user -> flex "sign contract(A: amount, D: down-payment) && deposit down-payment D"
            flex -> price_feed "get currency price"
            flex -> loans_vault "request loan (A-D)"
            loans_vault -> flex "send amount (A-D)"
            price_feed -> flex "push price update"
            timer -> flex "push time update"
            user -> flex "repay one or more times until pay-off the total of (A-D+I)"
            flex -> reserve_vault "forward payments total (A-D+I)"
            flex -> user "transfer ownership of A"
        }

        dynamic contracts "case1" "Flex liquidation" {
            title "Flex liquidation"
            price_feed -> flex "push price update"
            timer -> flex "push time update"
            flex -> reserve_vault "send the total amount A"
            autolayout
        }
        theme default
    }
    
}