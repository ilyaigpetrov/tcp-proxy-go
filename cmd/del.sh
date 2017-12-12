source <(head -2 redirect-traffic-to-proxy-client.sh)
sudo ip rule del fwmark 0x1 prohibit
