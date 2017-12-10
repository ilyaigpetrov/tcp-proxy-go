# if iptables are configured to route traffic to proxy then output traffic shouldn't by cycled into these proxies again.
sudo iptables -t nat -D OUTPUT -m tcp -p tcp --dport 80 -m owner --gid-owner hell -j RETURN
sudo iptables -t nat -D OUTPUT -m tcp -p tcp --dport 443 -m owner --gid-owner hell -j RETURN
sudo iptables -t nat -I OUTPUT -m tcp -p tcp --dport 80 -m owner --gid-owner hell -j RETURN
sudo iptables -t nat -I OUTPUT -m tcp -p tcp --dport 443 -m owner --gid-owner hell -j RETURN
sudo -g hell ./proxy-server $@
#sudo -u proxyrunner ./proxy-server $@
