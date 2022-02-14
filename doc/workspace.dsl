workspace {

    model {
        user = person "User"
        nolusenterprise = enterprise "Nolus" {
            
            nolus = softwareSystem "Nolus" {
                validatornode = group "Validator/Sentry Node" {
                    cosmosapp = container "Nolus Node" {
                        description "Tendermint PoS and Cosmos based blockchain"
                        technology "Cosmos SDK Application"
                        tax_agent = component "Tax Agent"
                        minter = component "Minter"
                        distributor = component "Distributor"
                        reserve_proxy = component "Reserve Vault Proxy"

                        tax_agent -> distributor "distribute the transaction gas"
                        tax_agent -> reserve_proxy "send extra fee"
                        minter -> distributor "mint amount on each block"
                    }
                    contracts = container "Smart Contracts" {
                        flex = component "Flex"
                        price_feed = component "Price Feed"
                        timer = component "Timer"
                        reserve_vault = component "Reserve Vault"
                        loans_vault = component "Loans Vault"
                        swap = component "Swap Gateway"
                        timer -> flex "time updates"
                        price_feed -> flex "price updates"
                        flex -> loans_vault "request amount"
                        flex -> reserve_vault "forward payments"
                        loans_vault -> swap "exchange"
                        reserve_vault -> swap "exchange"
                        reserve_vault -> loans_vault "rebalance"
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

                webapp = container "Web App" {
                    description "Static content and client-side rendered application"
                    technology "TypeScript, React"
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

        deploymentEnvironment dev {
            deploymentNode Nolus "description" "AWS" {
                deploymentNode Worker "description" "AWS EC2 dev-network-worker" {
                    deploymentNode "Validator Node" "" "nolusd" "" 3 {
                        containerInstance cosmosapp {}
                        properties {
                            p2p "tcp://127.0.0.1:26611, :26616, :26621"
                            rpc "http://0.0.0.0:26612, :26617, :26622"
                            api "http://0.0.0.0:26614, :26619, :26624"
                        }
                    }
                }
                webappHost = deploymentNode "WebApp Hosting" "" "AWS S3 web-dev.nolus.io" {
                    containerInstance webapp {}
                    url "https://app-dev.nolus.io"
                }
            }

            deploymentNode "Customer's device" "" "Desktop, laptop ot mobile" {
                clientBrowser = deploymentNode "Web Browser" "" "Chrome, Firefox, Safari" {
                    containerInstance webapp {}
                }
                webappHost -> clientBrowser "Delivers to the customer's web browser"
            }

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
            tax_agent -> reserve_proxy "send extra fee"
            reserve_proxy -> reserve_vault "send extra fee"
            tax_agent -> distributor "send remained gas"
            minter -> distributor "send newly minted coins"
            distributor -> distributor "distribute to the delegators"
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

        deployment nolus dev "nolus-dev-deployment" "Nolus Development Environment" {
            title "Nolus Development"
            include *
        }
        theme default
    }
    
}