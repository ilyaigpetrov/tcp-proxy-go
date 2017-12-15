# if iptables are configured to route traffic to proxy then output traffic shouldn't by cycled into these proxies again.
sudo groupadd proxyserver
sudo iptables -t nat -D OUTPUT -m tcp -p tcp --dport 80 -m owner --gid-owner proxyserver -j RETURN
sudo iptables -t nat -D OUTPUT -m tcp -p tcp --dport 443 -m owner --gid-owner proxyserver -j RETURN
sudo iptables -t nat -I OUTPUT -m tcp -p tcp --dport 80 -m owner --gid-owner proxyserver -j RETURN
sudo iptables -t nat -I OUTPUT -m tcp -p tcp --dport 443 -m owner --gid-owner proxyserver -j RETURN
sudo -g proxyserver ./ss -r $@
