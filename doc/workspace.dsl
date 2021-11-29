workspace {

    model {
        user = person "User"
        nolusenterprise = enterprise "Nolus" {
            
            nolus = softwareSystem "Nolus" {
                validatornode = group "Validator Node" {
                    cosmosapp = container "Cosmos App" {
                        bank = component "Bank"
                        wasm = component "Wasm Module"
                        oracle_module = component "Oracle Module"
                        ante = component "Ante Handlers"
                    }
                    contracts = container "Smart Contracts" {
                        flex = component "Flex"
                        price_data = component "Price Data"
                        scheduler_data = component "Scheduler Data"
                        loans_vault = component "Loans Vault"
                        nolus_vault = component "Nolus Vault"
                    }
                    cosmosapp -> contracts "Execute Trx"
                    contracts -> cosmosapp "Store State"
                }
                appserver = container "Application Server" {
                    -> cosmosapp "Source Events"
                    -> cosmosapp "Forward Transactions"
                }

                webapp = container "Web UI Client" {
                    -> appserver "Queries, Transactions"
                }

                oracle_feed = container "Oracle Feed" {
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

        component contracts {
            include *
        }

        dynamic cosmosapp fee_handler {
            title "Adding extra transaction fee"
            user -> ante "send transaction"
            ante -> nolus_vault "get vault Cosmos address"
            ante -> bank "send extra fee to vault Cosmos address"
        }

        dynamic cosmosapp oracle_msgs {
            title "Passing message to oracle smart contracts"
            oracle_feed -> ante "send price update"
            ante -> oracle_module "whitelist sender address"
            ante -> price_data "send without charging fee"
            price_data -> price_data "match msg sender address to whitelist"

            admin -> price_data "update whitelist"
            user -> oracle_module "update whitelists"
            oracle_module -> price_data "get and set whitelisted addresses"
        }

        dynamic contracts "case0" "all" {
            title "Flex successful close"
            user -> flex "sign contract(amount, down-payment) && deposit down-pay"
            flex -> price_data "get currency price"
            flex -> loans_vault "request loan"
            loans_vault -> flex "send amount/promise"
            user -> flex "repay one or more times until pay-off the total"
            flex -> user "transfer ownership"
            flex -> nolus_vault "send collateral"
            price_data -> flex "push price update"
            scheduler_data -> flex "push end time period notification"
            autolayout
        }

        dynamic contracts "case1" "loan payment in a single epoch" {
            title "Loan payment in a single epoch"
            user -> flex "sign contract(amount, down-payment) && deposit down-pay"
            flex -> price_data "get currency price"
            flex -> loans_vault "request loan"
            loans_vault -> flex "send amount/promise"
            user -> flex "repay one or more times until pay-off the total"
            flex -> user "transfer ownership"
            autolayout
        }

            dynamic contracts "case2" "update loans via oracles" {
            title "Update loans via oracles"
            price_data -> flex "push price update"
            scheduler_data -> flex "push end time period notification"
            flex -> nolus_vault "send collateral"
            autolayout
        }
        theme default
    }
    
}