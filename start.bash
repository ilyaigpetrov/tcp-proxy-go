sudo iptables -t nat -A OUTPUT -m tcp -p tcp --dport 80 -m owner --uid-owner proxyrunner -j RETURN
sudo iptables -t nat -A OUTPUT -m tcp -p tcp --dport 443 -m owner --uid-owner proxyrunner -j RETURN
sudo -u proxyrunner ./proxy
