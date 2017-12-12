# if iptables are configured to route traffic to proxy then output traffic shouldn't by cycled into these proxies again.
sudo useradd proxyclient
sudo iptables -t nat -D OUTPUT -m tcp -p tcp --dport 80 -m owner --uid-owner proxyclient -j RETURN
sudo iptables -t nat -D OUTPUT -m tcp -p tcp --dport 443 -m owner --uid-owner proxyclient -j RETURN
sudo iptables -t nat -I OUTPUT -m tcp -p tcp --dport 80 -m owner --uid-owner proxyclient -j RETURN
sudo iptables -t nat -I OUTPUT -m tcp -p tcp --dport 443 -m owner --uid-owner proxyclient -j RETURN

#sudo iptables -t filter -D INPUT -s 127.0.0.1 -p tcp --sport 443 -d 127.0.0.1 --dport 1111 -j REJECT


sudo -u proxyclient ./proxy-client -r $@
