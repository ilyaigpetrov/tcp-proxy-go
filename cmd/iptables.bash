sudo useradd proxyrunner
sudo iptables -t nat -A OUTPUT -m tcp -p tcp --dport 80 -m owner --uid-owner proxyrunner -j RETURN
sudo iptables -t nat -A OUTPUT -m tcp -p tcp --dport 443 -m owner --uid-owner proxyrunner -j RETURN
sudo iptables -t nat -A OUTPUT -p tcp -m tcp --dport 443 -j REDIRECT --to-ports 1111
sudo iptables -t nat -A OUTPUT -p tcp -m tcp --dport 80 -j REDIRECT --to-ports 1111
