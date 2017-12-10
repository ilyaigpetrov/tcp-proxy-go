# if iptables are configured to route traffic to proxy then output traffic shouldn't by cycled into these proxies again.
sudo useradd proxyclient
sudo iptables -t nat -A OUTPUT -m tcp -p tcp --dport 80 -m owner --uid-owner proxyclient -j RETURN
sudo iptables -t nat -A OUTPUT -m tcp -p tcp --dport 443 -m owner --uid-owner proxyclient -j RETURN
sudo -u proxyrunner ./proxy-client $@
