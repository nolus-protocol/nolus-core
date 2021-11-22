workspace {

    model {
        user = person "User"
        nolusenterprise = enterprise "Nolus" {
            
            nolus = softwareSystem "Nolus" {
                validatornode = group "Validator Node" {
                    cosmosapp = container "Cosmos App" {
                        bank = component "Bank"
                    }
                    contracts = container "Smart Contracts" {
                        flex = component "Flex"
                        dex = component "DEX"
                        loans_vault = component "Loans Vault"
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

        dynamic contracts {
            title "Flex successful close"
            user -> flex "sign contract(amount, down-payment)"
            user -> bank "deposit"
            user -> bank "down-pay"
            bank -> flex "send collateral"
            flex -> loans_vault "request loan"
            flex -> dex "buy target amount"
            user -> flex "repay one or more times until pay-off the total"
            flex -> bank "transfer ownership"
        }
        theme default
    }
    
}