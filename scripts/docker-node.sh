#!/bin/sh

CUSTOM_MONIKER=docker_node
DESTINATION=~/.nolus/config/genesis.json

# initialize node if no nolus folder found
if [ ! -d ~/.nolus ] ; then
   nolusd init $CUSTOM_MONIKER
fi

# set genesis.json 
if [ -f genesis.json ]; then
   cp genesis.json $DESTINATION
fi

# set persistent peers in config.toml
if [ -f genesis.json ]; then
   sed -i.bak -e "s~^persistent_peers *=.*~persistent_peers = \"$(cat persistent_peers.txt)\"~" ~/.nolus/config/config.toml
fi

# set the minimum gas price to 0.0025unls
sed -i "s|0stake|0.0025unls|g" ~/.nolus/config/app.toml