sudo iptables -t nat -D OUTPUT -p tcp -m tcp --dport 443 -j REDIRECT --to-ports 1111
sudo iptables -t nat -D OUTPUT -p tcp -m tcp --dport 80 -j REDIRECT --to-ports 1111
