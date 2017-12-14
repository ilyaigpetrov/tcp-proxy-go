sudo iptables -t nat -D OUTPUT -p tcp -m tcp --dport 80 -j REDIRECT --to-ports $1 -m owner ! --uid-owner proxyclient
sudo iptables -t nat -D OUTPUT -p tcp -m tcp --dport 443 -j REDIRECT --to-ports $1 -m owner ! --uid-owner proxyclient
sudo iptables -t nat -A OUTPUT -p tcp -m tcp --dport 80 -j REDIRECT --to-ports $1 -m owner ! --uid-owner proxyclient
sudo iptables -t nat -A OUTPUT -p tcp -m tcp --dport 443 -j REDIRECT --to-ports $1 -m owner ! --uid-owner proxyclient

#sudo iptables -t mangle -D OUTPUT -m tcp -p tcp --dport 2222 --sport 2222 -s 127.0.0.1 -d 127.0.0.1 -j MARK --set-mark 1
#sudo iptables -t mangle -I OUTPUT -m tcp -p tcp --dport 2222 --sport 2222 -s 127.0.0.1 -d 127.0.0.1 -j MARK --set-mark 1

#sudo ip rule del fwmark 0x1 prohibit
#sudo ip rule add fwmark 0x1 prohibit

#sudo iptables -t nat -D OUTPUT -p tcp -m tcp -j REDIRECT --dport 443 --to-ports $1 ! -s 127.0.0.1
#sudo iptables -t nat -D OUTPUT -p tcp -m tcp -j REDIRECT --dport 80 --to-ports $1 ! -s 127.0.0.1
#sudo iptables -t nat -I OUTPUT -p tcp -m tcp -j REDIRECT --dport 443 --to-ports $1 ! -s 127.0.0.1
#sudo iptables -t nat -I OUTPUT -p tcp -m tcp -j REDIRECT --dport 80 --to-ports $1 ! -s 127.0.0.1

#sudo iptables -t nat -D OUTPUT -p tcp -m tcp -j REDIRECT --dport 443 --to-ports $1 ! --sport $1 ! -s 127.0.0.1
#sudo iptables -t nat -D OUTPUT -p tcp -m tcp -j REDIRECT --dport 80 --to-ports $1 ! --sport $1 ! -s 127.0.0.1
#sudo iptables -t nat -I OUTPUT -p tcp -m tcp -j REDIRECT --dport 443 --to-ports $1 ! --sport $1 ! -s 127.0.0.1
#sudo iptables -t nat -I OUTPUT -p tcp -m tcp -j REDIRECT --dport 80 --to-ports $1 ! --sport $1 ! -s 127.0.0.1


