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
                        treasury_proxy = component "Treasury Proxy"

                        tax_agent -> distributor "distribute the transaction gas"
                        tax_agent -> treasury_proxy "send extra fee"
                        minter -> distributor "mint amount on each block"
                    }
                    contracts = container "Smart Contracts" {
                        price_oracle = component "Market Price Oracle" "Pair Prices and Alarms"
                        time_oracle = component "Global Time Oracle" "Time and Alarms"

                        leaser = component "Leaser" "Provide quotes, open new Lease"
                        lease = component "Lease" "Instance per Lease, holds the amount in crypto"
                        treasury = component "Treasury NLS" "Fees and profit"
                        stable_lpp = component "Liquidity Provider Pool LPN" "Liquidity Pool"
                        profit = component "Profit" "Collect swap tax and interest margin, and buy NLS back"
                        rewards = component "Rewards Distributor" "Calculate and distribute rewards to LPPs"
                        swap = component "Swap Gateway" "DEX interaction"

                        leaser -> stable_lpp "quote % interest rate"
                        leaser -> price_oracle "read pair prices"
                        leaser -> lease "open a new lease with % interest margin"
                        leaser -> lease "transfer the downpayment"
                        
                        time_oracle -> lease "time alarms"
                        price_oracle -> lease "price alarms"
                        lease -> stable_lpp "request loan LPN"
                        lease -> stable_lpp "repay loan LPN"
                        lease -> swap "exchange downpayment->crypto"
                        lease -> swap "exchange LPP loan LPN->crypto"
                        lease -> swap "exchange repayment->LPN"
                        lease -> profit "swap spread, interest margin LPN"
                        time_oracle -> profit "alarms on 48 hours"
                        profit -> swap "buy back NLS"
                        profit -> treasury "profit NLS"
                        rewards -> stable_lpp "rewards in NLS"
                        rewards -> treasury "get rewards"
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
                    market_data_feeder = component "Market Data Feeder" {
                       -> api_endpoint "Price updates"
                    }
                }
            }

            ibc = softwareSystem "Cosmos IBC Relay"

            ibc -> nolus "Relays-in"
            nolus -> ibc "Relays-out"

            faucet = softwareSystem "Faucet" {
                -> appserver "Send tx"
                faucetBackend = container "Faucet Backend" {
                    description "A faucet for Cosmos-SDK apps"
                    technology "Nolus CLI"
                    url "https://github.com/tendermint/faucet"
                }
                faucetUI = container "Faucet UI" {
                    description "A single page app"
                    technology "NodeJS"
                    url "https://github.com/scrtlabs/testnet-faucet"
                }
            }

            admin = person "Admin" {
                -> webapp "Uses"
            }
        }

        market_data_aggregator = softwareSystem "Market Data Aggregator" {
        }

        deploymentEnvironment dev {
            deploymentNode Nolus@AWS "description" "AWS" {
                deploymentNode Worker "description" "AWS EC2 dev-network-worker" {
                    validatorInstances = deploymentNode "Validator Node" "" "nolusd" "" 3 {
                        containerInstance cosmosapp {}
                    }
                }
                webappHost = deploymentNode "WebApp Hosting" "" "AWS S3 app-dev.nolus.io" {
                    containerInstance webapp {}
                    url "https://s3.eu-west-1.amazonaws.com:443/app-dev.nolus.io"
                }
                faucetNode = deploymentNode "Faucet" "" "AWS EC2 Faucet" {
                    faucetBackendNode = deploymentNode "Faucet Backend" "" "faucet JSON server" {
                        containerInstance faucetBackend {}
                    }
                    faucetUINode = deploymentNode "Faucet UI" "" "nodeJS server" {
                        containerInstance faucetUI {}
                        -> faucetBackendNode "Request test tokens" {
                            url "http://0.0.0.0:8000"
                        }
                    }
                }
                reverseProxyInstance = infrastructureNode reverseProxy "Proxy to the backend services and TLS termination" "HAProxy" {
                    -> faucetUINode "Plain JSON over HTTP" {
                        url "http://0.0.0.0:8080"
                    }
                    -> validatorInstances "Plain JSON over HTTP" {
                        properties {
                            rpc "http://0.0.0.0:26612, :26617, :26622"
                            api "http://0.0.0.0:26614, :26619, :26624"
                        }
                    }
                }
                cdn = infrastructureNode CDN "Content delivery for AWS S3 buckets, caching and protection" "CloudFront" {
                    -> webappHost "Load Nolus Web App" {
                        url "https://s3.eu-west-1.amazonaws.com:443/app-dev.nolus.io"
                    }
                }
                faucetBackendNode -> reverseProxyInstance "send tx" {
                    url "https://net-dev.nolus.io:26612"
                }
            }

            deploymentNode "CloudFlare" {
                infrastructureNode DNS "Domain Name Resolution of *.nolus.io to AWS EC2 public IPs" "CloudFlare" {}
                cloudFlareProxy = infrastructureNode Proxy "HTTP(S) Proxy with DDOS protection" "CloudFlare" {}
            }

            deploymentNode "Customer's device" "" "Desktop, laptop ot mobile" {
                clientBrowser = deploymentNode "Web Browser" "" "Chrome, Firefox, Safari" {
                    webappInstance = containerInstance webapp {}
                    containerInstance faucetUI {}
                }
                clientBrowser -> cdn "Web-app Hosting" "Load Nolus Web App" {
                    url "https://app-dev.nolus.io:443"
                }
                clientBrowser -> reverseProxyInstance "JSON Queries and transactions to HTTPS rpc&api endpoints" {
                    properties {
                        rpc "https://net-dev.nolus.io:26612, :26617, :26622"
                        api "https://net-dev.nolus.io:26614, :26619, :26624"
                    }   
                }
                clientBrowser -> reverseProxyInstance "Load Faucet app and send JSON test tokens requests" {
                    url "https://faucet-dev.nolus.io:443"
                }
            }
        }
        market_data_feeder -> market_data_aggregator "Fetch Data"

        user -> webapp "Uses"
        user -> faucet "Request Test Coins"
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
            tax_agent -> treasury_proxy "send extra fee"
            treasury_proxy -> treasury "send extra fee"
            tax_agent -> distributor "send remained gas"
            minter -> distributor "send newly minted coins"
            distributor -> distributor "distribute to the delegators"
        }

        dynamic contracts oracle_msgs {
            title "Price Feeds"

            admin -> price_oracle "manage supported price pairs"
            admin -> price_oracle "manage whitelisted operators"

            market_data_feeder -> price_oracle "read supported price pairs"
            market_data_feeder -> market_data_aggregator "fetch data"
            market_data_feeder -> price_oracle "price updates"
            
            market_data_feeder -> price_oracle "send observations"
            price_oracle -> price_oracle "match msg sender address to whitelist"
            price_oracle -> price_oracle "a price pair get ready if some number of feeders have been active for some period now"
            leaser -> price_oracle "read pair prices"
            lease -> price_oracle "read pair prices"
            price_oracle -> lease "push price alerts"
        }

        dynamic contracts "case0" "Lease successful close" {
            title "Lease successful close"
            user -> leaser "quote lease(C: currency, D: down-payment)"
            leaser -> stable_lpp "quote % interest rate"
            leaser -> price_oracle "read C/LPN, D/LPN prices"
            leaser -> user "quote(A: amount in C, B: borrowed amount in LPN, I: interest rate)"

            user -> leaser "open lease(C: currency) & transfer(D: down-payment)"
            leaser -> lease "instantiate(C, D, Ii: % interest margin)"
            leaser -> lease "transfer(D)"
            lease -> swap "exchange D->C"
            lease -> price_oracle "read C/LPN, D/LPN prices"
            lease -> stable_lpp "request (A-D) loan LPN"
            lease -> swap "exchange LPN->C"
            leaser -> user "open lease done"

            user -> lease "repay() & transfer(R: re-payment)"
            lease -> swap "exchange R->LPN"
            lease -> stable_lpp "repay loan+LPP interest LPN"
            lease -> profit "transfer(interest margin + swap spread)"
            lease -> user "transfer(A)"
        }

        dynamic contracts "case1" "Lease liquidation" {
            title "Lease partial liquidation"
            price_oracle -> lease "push price alerts"
            time_oracle -> lease "push time alerts"
            lease -> swap "exchange some C->LPN"
            lease -> lease "decrease A"
            lease -> stable_lpp "repay loan+LPP interest LPN"
            lease -> profit "transfer(interest margin + swap spread)"
        }

        deployment * dev "nolus-dev-deployment" "Nolus Development Environment" {
            title "Nolus Development"
            include *
        }
        theme default
    }
    
}