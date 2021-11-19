workspace {

    model {
        user = person "User"
        nolusenterprise = enterprise "Nolus" {
            
            nolus = softwareSystem "Nolus" {
                validatornode = group "Validator Node" {
                    cosmosapp = container "Cosmos App"
                    contracts = container "Smart Contracts"
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
        
        theme default
    }
    
}